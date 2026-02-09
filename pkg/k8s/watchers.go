package k8s

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

// WatchNamespaceResources monitors resources for a specific namespace
func WatchNamespaceResources(ctx context.Context, writeChan chan<- []byte, namespace string) {
	gvrs := []schema.GroupVersionResource{
		{Group: "", Version: "v1", Resource: "pods"},
		{Group: "", Version: "v1", Resource: "services"},
		{Group: "apps", Version: "v1", Resource: "deployments"},
	}

	var wg sync.WaitGroup
	for _, gvr := range gvrs {
		wg.Add(1)
		time.Sleep(50 * time.Millisecond) // Stagger start to be gentle on APIServer

		go func(gvr schema.GroupVersionResource) {
			defer wg.Done()
			watchAndSend(ctx, DynamicClient, gvr, namespace, writeChan)
		}(gvr)
	}

	// Wait for all watchers to finish (via context cancel) then close channel
	go func() {
		<-ctx.Done()
		wg.Wait()
		close(writeChan)
	}()
}

// WatchUserNamespaceResources monitors resource changes in a single namespace
func WatchUserNamespaceResources(ctx context.Context, namespace string, writeChan chan<- []byte) {
	gvrs := []schema.GroupVersionResource{
		{Group: "", Version: "v1", Resource: "pods"},
		{Group: "", Version: "v1", Resource: "services"},
		{Group: "apps", Version: "v1", Resource: "deployments"},
	}

	var wg sync.WaitGroup
	for _, gvr := range gvrs {
		wg.Add(1)
		go func(gvr schema.GroupVersionResource) {
			defer wg.Done()
			watchUserAndSend(ctx, namespace, gvr, writeChan)
		}(gvr)
	}

	// Wait logic handled by caller or context
	wg.Wait()
}

func watchUserAndSend(ctx context.Context, namespace string, gvr schema.GroupVersionResource, writeChan chan<- []byte) {
	// lastSnapshot holds last sent status signature per resource name
	lastSnapshot := make(map[string]string)

	sendObject := func(eventType string, obj *unstructured.Unstructured) error {
		name := obj.GetName()

		// Always send deletes
		if eventType != "DELETED" {
			// Compute compact snapshot to decide whether to send
			snap := statusSnapshotString(obj)
			if prev, ok := lastSnapshot[name]; ok {
				if prev == snap {
					// No meaningful status change, skip sending
					return nil
				}
			}
			lastSnapshot[name] = snap
		} else {
			// remove snapshot on delete
			delete(lastSnapshot, name)
		}

		data := buildDataMap(eventType, obj)
		msg, err := json.Marshal(data)
		if err != nil {
			return err
		}

		select {
		case writeChan <- msg:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Prevent blocking if client is slow
			return fmt.Errorf("client buffer full, dropping message for %s", name)
		}
	}

	list, err := DynamicClient.Resource(gvr).Namespace(namespace).List(ctx, metav1.ListOptions{})
	if err == nil {
		for _, item := range list.Items {
			_ = sendObject("ADDED", &item)
		}
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Second * 30):
			// Simple reconnection logic
			watcher, err := DynamicClient.Resource(gvr).Namespace(namespace).Watch(ctx, metav1.ListOptions{})
			if err != nil {
				continue
			}

			func() {
				defer watcher.Stop()
				for {
					select {
					case <-ctx.Done():
						return
					case event, ok := <-watcher.ResultChan():
						if !ok {
							return
						}
						if obj, ok := event.Object.(*unstructured.Unstructured); ok {
							_ = sendObject(string(event.Type), obj)
						}
					}
				}
			}()
		}
	}
}

func watchAndSend(
	ctx context.Context,
	dynClient dynamic.Interface,
	gvr schema.GroupVersionResource,
	ns string,
	writeChan chan<- []byte,
) {
	// lastSnapshot holds last sent status signature per resource name
	lastSnapshot := make(map[string]string)

	sendObject := func(eventType string, obj *unstructured.Unstructured) error {
		name := obj.GetName()

		if eventType != "DELETED" {
			snap := statusSnapshotString(obj)
			if prev, ok := lastSnapshot[name]; ok {
				if prev == snap {
					return nil
				}
			}
			lastSnapshot[name] = snap
		} else {
			delete(lastSnapshot, name)
		}

		data := buildDataMap(eventType, obj)
		msg, err := json.Marshal(data)
		if err != nil {
			return err
		}

		select {
		case writeChan <- msg:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	// Initial List
	list, err := dynClient.Resource(gvr).Namespace(ns).List(ctx, metav1.ListOptions{})
	if err == nil {
		for _, item := range list.Items {
			if err := sendObject("ADDED", &item); err != nil && ctx.Err() != context.Canceled {
				slog.Warn("Failed to send list item", slog.Any("error", err))
			}
		}
	} else if ctx.Err() == context.Canceled {
		return
	} else {
		slog.Error("List error for resource",
			slog.String("resource", gvr.Resource),
			slog.String("group", gvr.Group),
			slog.Any("error", err))
	}

	// Watch Loop
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		watcher, err := dynClient.Resource(gvr).Namespace(ns).Watch(ctx, metav1.ListOptions{})
		if err != nil {
			if ctx.Err() == context.Canceled {
				return
			}
			time.Sleep(5 * time.Second)
			continue
		}

		func() {
			defer watcher.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case event, ok := <-watcher.ResultChan():
					if !ok {
						return
					}

					obj, ok := event.Object.(*unstructured.Unstructured)
					if !ok {
						continue
					}

					if err := sendObject(string(event.Type), obj); err != nil && ctx.Err() != context.Canceled {
						slog.Warn("Failed to send watch event", slog.Any("error", err))
					}
				}
			}
		}()
	}
}
