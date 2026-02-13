package k8s

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/linskybing/platform-go/internal/config"
	schedulingv1 "k8s.io/api/scheduling/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func EnsurePriorityClass(ctx context.Context) error {
	if config.ConfigFilePriorityClassName == "" {
		return nil
	}
	if Clientset == nil {
		slog.Debug("[MOCK] ensure priority class", "name", config.ConfigFilePriorityClassName)
		return nil
	}
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := Clientset.SchedulingV1().PriorityClasses().Get(ctx, config.ConfigFilePriorityClassName, metav1.GetOptions{})
	if err == nil {
		return nil
	}
	if !apierrors.IsNotFound(err) {
		return fmt.Errorf("get priority class: %w", err)
	}

	pc := &schedulingv1.PriorityClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: config.ConfigFilePriorityClassName,
		},
		Value:         int32(config.ConfigFilePriorityValue),
		GlobalDefault: false,
		Description:   "Priority class for configfile pods",
	}

	_, err = Clientset.SchedulingV1().PriorityClasses().Create(ctx, pc, metav1.CreateOptions{})
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return fmt.Errorf("create priority class: %w", err)
	}
	return nil
}
