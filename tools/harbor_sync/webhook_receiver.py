from flask import Flask, request, jsonify
import os
import logging
import psycopg2
from datetime import datetime

logging.basicConfig(level=logging.INFO)
app = Flask(__name__)

DB_DSN = os.environ.get('DATABASE_URL')  # postgres://user:pass@host:port/dbname
PG_USER = os.environ.get('PG_USER')
PG_PASSWORD = os.environ.get('PG_PASSWORD')
PG_DB = os.environ.get('PG_DB', 'platform')

WEBHOOK_SECRET = os.environ.get('WEBHOOK_SECRET')


def get_conn():
    dsn = DB_DSN
    if not dsn:
        if not (PG_USER and PG_PASSWORD):
            raise RuntimeError('DATABASE_URL or PG_USER/PG_PASSWORD must be set')
        dsn = f'postgres://{PG_USER}:{PG_PASSWORD}@postgres:5432/{PG_DB}'
    return psycopg2.connect(dsn)


def _get_or_create_repo(cur, full_name):
    cur.execute('SELECT id FROM repos WHERE full_name=%s', (full_name,))
    row = cur.fetchone()
    if row:
        return row[0]
    cur.execute(
        "INSERT INTO repos (full_name) VALUES (%s) ON CONFLICT (full_name) DO NOTHING RETURNING id",
        (full_name,),
    )
    row = cur.fetchone()
    if row:
        return row[0]
    cur.execute('SELECT id FROM repos WHERE full_name=%s', (full_name,))
    row = cur.fetchone()
    return row[0] if row else None


def _get_or_create_tag(cur, repo_id, tag):
    cur.execute('SELECT id FROM tags WHERE repository_id=%s AND tag=%s', (repo_id, tag))
    row = cur.fetchone()
    if row:
        return row[0]
    cur.execute(
        "INSERT INTO tags (repository_id, tag) VALUES (%s, %s) ON CONFLICT (repository_id, tag) DO NOTHING RETURNING id",
        (repo_id, tag),
    )
    row = cur.fetchone()
    if row:
        return row[0]
    cur.execute('SELECT id FROM tags WHERE repository_id=%s AND tag=%s', (repo_id, tag))
    row = cur.fetchone()
    return row[0] if row else None


def mark_image_pulled(repository, tag):
    now = datetime.utcnow()
    tag_id = None
    with get_conn() as conn:
        with conn.cursor() as cur:
            repo_id = _get_or_create_repo(cur, repository)
            if repo_id is None:
                raise RuntimeError('failed to ensure repo')
            tag_id = _get_or_create_tag(cur, repo_id, tag)
            if tag_id is None:
                raise RuntimeError('failed to ensure tag')

            cur.execute(
                'UPDATE image_pulls SET is_pulled=true, last_pulled_at=%s WHERE tag_id=%s', (now, tag_id)
            )
            if cur.rowcount == 0:
                cur.execute(
                    'INSERT INTO image_pulls (tag_id, is_pulled, last_pulled_at) VALUES (%s, true, %s)',
                    (tag_id, now),
                )

            cur.execute(
                'UPDATE allowed_images SET tag_id=%s, raw_name=%s, raw_tag=%s WHERE name=%s AND tag=%s',
                (tag_id, repository, tag, repository, tag),
            )
        conn.commit()
    logging.info('Marked pulled: %s:%s (tag_id=%s)', repository, tag, tag_id)


def mark_image_deleted(repository, tag=None):
    now = datetime.utcnow()
    with get_conn() as conn:
        with conn.cursor() as cur:
            if tag:
                cur.execute(
                    '''UPDATE allowed_images SET deleted_at=%s, harbor_deleted=true, harbor_deleted_at=%s
                               WHERE name=%s AND tag=%s''',
                    (now, now, repository, tag),
                )
            else:
                cur.execute(
                    '''UPDATE allowed_images SET deleted_at=%s, harbor_deleted=true, harbor_deleted_at=%s
                               WHERE name=%s''',
                    (now, now, repository),
                )

            cur.execute('SELECT id FROM repos WHERE full_name=%s', (repository,))
            row = cur.fetchone()
            if row:
                repo_id = row[0]
                if tag:
                    cur.execute('SELECT id FROM tags WHERE repository_id=%s AND tag=%s', (repo_id, tag))
                    trow = cur.fetchone()
                    if trow:
                        tag_id = trow[0]
                        cur.execute('UPDATE image_pulls SET is_pulled=false WHERE tag_id=%s', (tag_id,))
                else:
                    cur.execute(
                        '''UPDATE image_pulls SET is_pulled=false WHERE tag_id IN (SELECT id FROM tags WHERE repository_id=%s)''',
                        (repo_id,),
                    )
        conn.commit()
    logging.info('Marked deleted in DB: %s:%s', repository, tag)


def parse_harbor_payload(payload):
    results = []
    event_data = payload.get('event_data') or payload.get('artifact') or payload

    resource = event_data.get('resource') if isinstance(event_data, dict) else None
    if resource:
        repo = resource.get('repository') or resource.get('repo') or payload.get('repository', {}).get('name')
        tags = []
        if 'tag' in resource and resource.get('tag'):
            tags = [resource.get('tag')]
        elif 'tags' in resource and isinstance(resource.get('tags'), list):
            tags = [t.get('name') if isinstance(t, dict) else t for t in resource.get('tags')]
        if repo and tags:
            for t in tags:
                results.append((repo, t))
            return results

    repo = payload.get('repository', {}).get('name')
    push_data = payload.get('push_data') or payload.get('push') or {}
    tag = push_data.get('tag')
    if repo and tag:
        results.append((repo, tag))
        return results

    if isinstance(payload.get('artifact'), dict):
        art = payload['artifact']
        repo = art.get('repository') or payload.get('repository', {}).get('name')
        if 'tags' in art:
            tags = [t.get('name') if isinstance(t, dict) else t for t in art.get('tags')]
            for t in tags:
                results.append((repo, t))
    return results


def parse_harbor_delete(payload):
    results = []
    if isinstance(payload.get('artifact'), dict):
        art = payload['artifact']
        repo = art.get('repository') or payload.get('repository', {}).get('name')
        tags = art.get('tags') or []
        if tags:
            for t in tags:
                tag_name = t.get('name') if isinstance(t, dict) else t
                results.append((repo, tag_name))
            return results
        if repo:
            results.append((repo, None))
            return results

    repo = payload.get('repository', {}).get('name')
    if repo:
        push_data = payload.get('push_data') or payload.get('push') or {}
        tag = push_data.get('tag')
        results.append((repo, tag))
    return results


@app.route('/webhook', methods=['POST'])
def webhook():
    token = request.headers.get('X-Webhook-Token')
    if WEBHOOK_SECRET:
        if not token or token != WEBHOOK_SECRET:
            return jsonify({'error': 'unauthorized'}), 401

    try:
        payload = request.get_json(force=True)
    except Exception:
        return jsonify({'error': 'invalid json'}), 400

    evtype = (payload.get('type') or payload.get('event_type') or payload.get('event_type_lite') or '').lower()
    is_delete = False
    if 'delete' in evtype or ('artifact' in evtype and 'delete' in evtype):
        is_delete = True
    if payload.get('operation') and payload.get('operation').upper() == 'DELETE':
        is_delete = True

    if is_delete:
        del_pairs = parse_harbor_delete(payload)
        if not del_pairs:
            logging.warning('No delete info found in payload')
            return jsonify({'status': 'ignored'}), 200
        for repo, tag in del_pairs:
            try:
                mark_image_deleted(repo, tag)
            except Exception:
                logging.exception('DB delete mark failed for %s:%s', repo, tag)
                return jsonify({'error': 'db error'}), 500
        return jsonify({'status': 'ok', 'deleted': len(del_pairs)}), 200

    pairs = parse_harbor_payload(payload)
    if not pairs:
        logging.warning('No repo:tag found in payload')
        return jsonify({'status': 'ignored'}), 200

    for repo, tag in pairs:
        try:
            mark_image_pulled(repo, tag)
        except Exception:
            logging.exception('DB update failed for %s:%s', repo, tag)
            return jsonify({'error': 'db error'}), 500

    return jsonify({'status': 'ok', 'updated': len(pairs)}), 200


@app.route('/healthz', methods=['GET'])
def healthz():
    return jsonify({'status': 'ok'}), 200


if __name__ == '__main__':
    app.run(host='0.0.0.0', port=int(os.environ.get('PORT', 8080)))

from datetime import datetime

logging.basicConfig(level=logging.INFO)
app = Flask(__name__)

DB_DSN = os.environ.get('DATABASE_URL')  # postgres://user:pass@host:port/dbname
PG_USER = os.environ.get('PG_USER')
PG_PASSWORD = os.environ.get('PG_PASSWORD')
PG_DB = os.environ.get('PG_DB', 'platform')

WEBHOOK_SECRET = os.environ.get('WEBHOOK_SECRET')

def get_conn():
    dsn = DB_DSN
    if not dsn:
        if not (PG_USER and PG_PASSWORD):
            raise RuntimeError('DATABASE_URL or PG_USER/PG_PASSWORD must be set')
        dsn = f'postgres://{PG_USER}:{PG_PASSWORD}@postgres:5432/{PG_DB}'
