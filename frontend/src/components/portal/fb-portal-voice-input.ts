/**
 * fb-portal-voice-input — Voice dictation with offline MediaRecorder fallback.
 *
 * Online: webkitSpeechRecognition for live transcription.
 * Offline: MediaRecorder captures audio blobs, queues in IndexedDB.
 *
 * Phase 18: See FRONTEND_SCOPE.md §15.2 (Voice-First Field Portal)
 */
import { html, css, nothing } from 'lit';
import { customElement, state } from 'lit/decorators.js';
import { FBElement } from '../base/FBElement';
import { OfflineQueue } from '../../services/offline-queue';

// Web Speech API types (not in standard DOM lib)
interface SpeechRecognitionEvent extends Event {
    readonly results: SpeechRecognitionResultList;
    readonly resultIndex: number;
}

interface SpeechRecognitionResultList {
    readonly length: number;
    item(index: number): SpeechRecognitionResult;
    [index: number]: SpeechRecognitionResult;
}

interface SpeechRecognitionResult {
    readonly length: number;
    readonly isFinal: boolean;
    item(index: number): SpeechRecognitionAlternative;
    [index: number]: SpeechRecognitionAlternative;
}

interface SpeechRecognitionAlternative {
    readonly transcript: string;
    readonly confidence: number;
}

interface SpeechRecognition extends EventTarget {
    continuous: boolean;
    interimResults: boolean;
    lang: string;
    onstart: ((this: SpeechRecognition, ev: Event) => void) | null;
    onresult: ((this: SpeechRecognition, ev: SpeechRecognitionEvent) => void) | null;
    onend: ((this: SpeechRecognition, ev: Event) => void) | null;
    onerror: ((this: SpeechRecognition, ev: Event) => void) | null;
    start(): void;
    stop(): void;
    abort(): void;
}

type VoiceState = 'idle' | 'recording' | 'processing' | 'offline';

@customElement('fb-portal-voice-input')
export class FBPortalVoiceInput extends FBElement {
    static override styles = [
        FBElement.styles,
        css`
            :host {
                display: flex;
                flex-direction: column;
                align-items: center;
                padding: 16px;
            }

            .mic-btn {
                width: 72px;
                height: 72px;
                border-radius: 50%;
                border: none;
                cursor: pointer;
                display: flex;
                align-items: center;
                justify-content: center;
                transition: all 0.2s;
                position: relative;
            }

            .mic-btn.idle {
                background: var(--fb-accent, #00FFA3);
                color: #0A0B10;
            }

            .mic-btn.recording {
                background: var(--fb-accent, #00FFA3);
                color: #0A0B10;
                animation: pulse 1.5s ease-in-out infinite;
            }

            .mic-btn.processing {
                background: var(--fb-surface-2, #1E2029);
                color: var(--fb-text-secondary, #8B8D98);
                cursor: wait;
            }

            .mic-btn.offline {
                background: rgba(245, 158, 11, 0.2);
                color: #f59e0b;
            }

            @keyframes pulse {
                0%, 100% { box-shadow: 0 0 0 0 rgba(0, 255, 163, 0.4); }
                50% { box-shadow: 0 0 0 16px rgba(0, 255, 163, 0); }
            }

            .mic-icon {
                width: 28px;
                height: 28px;
            }

            .status-text {
                margin-top: 8px;
                font-size: 13px;
                color: var(--fb-text-secondary, #8B8D98);
            }

            .transcript {
                margin-top: 12px;
                padding: 12px 16px;
                background: var(--fb-surface-1, #161821);
                border: 1px solid var(--fb-border, rgba(255,255,255,0.05));
                border-radius: 8px;
                font-size: 14px;
                color: var(--fb-text-primary, #F0F0F5);
                width: 100%;
                max-width: 400px;
                min-height: 40px;
            }

            .offline-badge {
                margin-top: 8px;
                padding: 4px 10px;
                border-radius: 4px;
                background: rgba(245, 158, 11, 0.1);
                color: #f59e0b;
                font-size: 12px;
                font-weight: 600;
            }
        `,
    ];

    @state() private _state: VoiceState = 'idle';
    @state() private _transcript = '';
    @state() private _isOnline = navigator.onLine;
    @state() private _pendingCount = 0;

    private _recognition: SpeechRecognition | null = null;
    private _recorder: MediaRecorder | null = null;
    private _audioChunks: Blob[] = [];
    private _queue = new OfflineQueue();

    override connectedCallback() {
        super.connectedCallback();
        window.addEventListener('online', this._handleOnline);
        window.addEventListener('offline', this._handleOffline);
        this._updatePendingCount();
    }

    override disconnectedCallback() {
        super.disconnectedCallback();
        window.removeEventListener('online', this._handleOnline);
        window.removeEventListener('offline', this._handleOffline);
        this._stopRecording();
    }

    private _handleOnline = () => {
        this._isOnline = true;
        if (this._state === 'offline') this._state = 'idle';
        this._queue.syncPending();
    };

    private _handleOffline = () => {
        this._isOnline = false;
    };

    private async _updatePendingCount() {
        this._pendingCount = await this._queue.getPendingCount();
    }

    private _handleMicClick() {
        if (this._state === 'recording') {
            this._stopRecording();
            return;
        }

        if (this._state === 'processing') return;

        if (this._isOnline && 'webkitSpeechRecognition' in window) {
            this._startSpeechRecognition();
        } else {
            this._startMediaRecorder();
        }
    }

    private _startSpeechRecognition() {
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        const SpeechRecognition = (window as any).webkitSpeechRecognition || (window as any).SpeechRecognition;
        if (!SpeechRecognition) {
            this._startMediaRecorder();
            return;
        }

        const recognition: SpeechRecognition = new SpeechRecognition();
        this._recognition = recognition;
        recognition.continuous = true;
        recognition.interimResults = true;
        recognition.lang = 'en-US';

        recognition.onstart = () => {
            this._state = 'recording';
            this._transcript = '';
        };

        recognition.onresult = (event: SpeechRecognitionEvent) => {
            let transcript = '';
            for (let i = 0; i < event.results.length; i++) {
                const result = event.results[i];
                if (result) {
                    const alt = result[0];
                    if (alt) transcript += alt.transcript;
                }
            }
            this._transcript = transcript;
        };

        recognition.onend = () => {
            this._state = 'idle';
            if (this._transcript) {
                this._emitVoiceInput(this._transcript, undefined, false);
            }
        };

        recognition.onerror = () => {
            this._state = 'idle';
        };

        recognition.start();
    }

    private async _startMediaRecorder() {
        try {
            const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
            this._recorder = new MediaRecorder(stream);
            this._audioChunks = [];

            this._recorder.ondataavailable = (e) => {
                if (e.data.size > 0) {
                    this._audioChunks.push(e.data);
                }
            };

            this._recorder.onstop = async () => {
                const audioBlob = new Blob(this._audioChunks, { type: 'audio/webm' });
                stream.getTracks().forEach(t => t.stop());

                if (!this._isOnline) {
                    await this._queue.queueVoiceMemo(audioBlob, { capturedAt: new Date().toISOString() });
                    await this._updatePendingCount();
                    this._state = 'idle';
                    this._transcript = '(Saved offline — will sync when connected)';
                } else {
                    this._state = 'processing';
                    this._emitVoiceInput('', audioBlob, false);
                    this._state = 'idle';
                }
            };

            this._state = this._isOnline ? 'recording' : 'offline';
            this._recorder.start();
        } catch {
            this._state = 'idle';
        }
    }

    private _stopRecording() {
        if (this._recognition) {
            this._recognition.stop();
            this._recognition = null;
        }
        if (this._recorder && this._recorder.state === 'recording') {
            this._recorder.stop();
            this._recorder = null;
        }
    }

    private _emitVoiceInput(text: string, audioBlob: Blob | undefined, isOffline: boolean) {
        this.dispatchEvent(new CustomEvent('voice-input', {
            detail: { text, audioBlob, isOffline },
            bubbles: true,
            composed: true,
        }));
    }

    override render() {
        return html`
            <button class="mic-btn ${this._state}" @click=${this._handleMicClick}>
                ${this._state === 'processing'
                    ? html`<svg class="mic-icon" viewBox="0 0 24 24" fill="currentColor"><circle cx="12" cy="12" r="3"/><path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm0 18c-4.42 0-8-3.58-8-8s3.58-8 8-8 8 3.58 8 8-3.58 8-8 8z" opacity="0.3"><animateTransform attributeName="transform" type="rotate" from="0 12 12" to="360 12 12" dur="1s" repeatCount="indefinite"/></path></svg>`
                    : html`<svg class="mic-icon" viewBox="0 0 24 24" fill="currentColor"><path d="M12 14c1.66 0 3-1.34 3-3V5c0-1.66-1.34-3-3-3S9 3.34 9 5v6c0 1.66 1.34 3 3 3zm-1-9c0-.55.45-1 1-1s1 .45 1 1v6c0 .55-.45 1-1 1s-1-.45-1-1V5z"/><path d="M17 11c0 2.76-2.24 5-5 5s-5-2.24-5-5H5c0 3.53 2.61 6.43 6 6.92V21h2v-3.08c3.39-.49 6-3.39 6-6.92h-2z"/></svg>`}
            </button>

            <span class="status-text">
                ${this._state === 'idle' ? 'Tap to speak' : ''}
                ${this._state === 'recording' ? 'Listening...' : ''}
                ${this._state === 'processing' ? 'Processing...' : ''}
                ${this._state === 'offline' ? 'Recording offline...' : ''}
            </span>

            ${this._transcript ? html`<div class="transcript">${this._transcript}</div>` : nothing}
            ${!this._isOnline ? html`<span class="offline-badge">Offline${this._pendingCount > 0 ? ` (${this._pendingCount} pending)` : ''}</span>` : nothing}
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'fb-portal-voice-input': FBPortalVoiceInput;
    }
}
