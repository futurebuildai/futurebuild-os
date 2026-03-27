/**
 * OfflineQueue — IndexedDB-backed offline queue for voice memos.
 *
 * Phase 18: See FRONTEND_SCOPE.md §15.2 (Voice-First Field Portal)
 *
 * Stores audio blobs in IndexedDB when offline.
 * Auto-syncs when connection is restored.
 */

const DB_NAME = 'fb-offline-queue';
const DB_VERSION = 1;
const STORE_NAME = 'voice-memos';

interface VoiceMemoEntry {
    id?: number;
    audioBlob: Blob;
    metadata: Record<string, string>;
    createdAt: string;
}

export class OfflineQueue {
    private _db: IDBDatabase | null = null;
    private _syncing = false;

    constructor() {
        if (typeof window !== 'undefined') {
            window.addEventListener('online', () => this.syncPending());
        }
    }

    private _openDB(): Promise<IDBDatabase> {
        if (this._db) return Promise.resolve(this._db);

        return new Promise((resolve, reject) => {
            const request = indexedDB.open(DB_NAME, DB_VERSION);

            request.onupgradeneeded = () => {
                const db = request.result;
                if (!db.objectStoreNames.contains(STORE_NAME)) {
                    db.createObjectStore(STORE_NAME, { keyPath: 'id', autoIncrement: true });
                }
            };

            request.onsuccess = () => {
                this._db = request.result;
                resolve(this._db);
            };

            request.onerror = () => {
                reject(request.error);
            };
        });
    }

    async queueVoiceMemo(audioBlob: Blob, metadata: Record<string, string>): Promise<void> {
        const db = await this._openDB();
        return new Promise((resolve, reject) => {
            const tx = db.transaction(STORE_NAME, 'readwrite');
            const store = tx.objectStore(STORE_NAME);
            const entry: VoiceMemoEntry = {
                audioBlob,
                metadata,
                createdAt: new Date().toISOString(),
            };
            const req = store.add(entry);
            req.onsuccess = () => resolve();
            req.onerror = () => reject(req.error);
        });
    }

    async getPendingCount(): Promise<number> {
        try {
            const db = await this._openDB();
            return new Promise((resolve, reject) => {
                const tx = db.transaction(STORE_NAME, 'readonly');
                const store = tx.objectStore(STORE_NAME);
                const req = store.count();
                req.onsuccess = () => resolve(req.result);
                req.onerror = () => reject(req.error);
            });
        } catch {
            return 0;
        }
    }

    async syncPending(): Promise<void> {
        if (this._syncing || !navigator.onLine) return;
        this._syncing = true;

        try {
            const db = await this._openDB();
            const entries = await this._getAllEntries(db);

            for (const entry of entries) {
                try {
                    await this._uploadVoiceMemo(entry);
                    await this._deleteEntry(db, entry.id!);
                } catch (err) {
                    console.warn('[OfflineQueue] Failed to sync memo:', err);
                    break;
                }
            }
        } finally {
            this._syncing = false;
        }
    }

    private _getAllEntries(db: IDBDatabase): Promise<VoiceMemoEntry[]> {
        return new Promise((resolve, reject) => {
            const tx = db.transaction(STORE_NAME, 'readonly');
            const store = tx.objectStore(STORE_NAME);
            const req = store.getAll();
            req.onsuccess = () => resolve(req.result as VoiceMemoEntry[]);
            req.onerror = () => reject(req.error);
        });
    }

    private _deleteEntry(db: IDBDatabase, id: number): Promise<void> {
        return new Promise((resolve, reject) => {
            const tx = db.transaction(STORE_NAME, 'readwrite');
            const store = tx.objectStore(STORE_NAME);
            const req = store.delete(id);
            req.onsuccess = () => resolve();
            req.onerror = () => reject(req.error);
        });
    }

    private async _uploadVoiceMemo(entry: VoiceMemoEntry): Promise<void> {
        const formData = new FormData();
        formData.append('audio', entry.audioBlob, 'voice-memo.webm');
        formData.append('metadata', JSON.stringify(entry.metadata));

        const response = await fetch('/api/v1/portal/voice-memos', {
            method: 'POST',
            body: formData,
        });

        if (!response.ok) {
            throw new Error(`Upload failed: ${response.status}`);
        }
    }
}
