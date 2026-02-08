/**
 * FBPhotoGallery - Project Asset Gallery with Vision Status Badges
 * See STEP_85_VISION_BADGES.md Section 1
 *
 * Displays project photo assets in a thumbnail grid with overlayed
 * vision verification badges. Each thumbnail shows the AI analysis
 * status (Verifying, Verified, Flagged, Failed).
 *
 * @fires fb-asset-click - When a thumbnail is clicked (detail: { asset })
 * @fires fb-badge-click - Bubbled from fb-vision-badge (detail: { status, summary })
 */
import { html, css, TemplateResult } from 'lit';
import { customElement, property } from 'lit/decorators.js';
import { FBElement } from '../base/FBElement';
import type { VisionBadgeStatus } from './fb-vision-badge';
import './fb-vision-badge';

/**
 * Asset data for gallery display.
 * Matches the GET /api/v1/projects/:id/assets response shape.
 */
export interface GalleryAsset {
    id: string;
    file_name: string;
    file_url: string;
    analysis_status: VisionBadgeStatus;
    analysis_summary?: string;
}

/**
 * Maps database analysis_status values to badge status values.
 * DB: processing → verifying, completed → verified, failed → failed
 * The "flagged" status comes from analysis_result containing flag data.
 */
function mapAnalysisStatus(dbStatus: string, hasFlagData: boolean): VisionBadgeStatus {
    if (dbStatus === 'completed' && hasFlagData) return 'flagged';
    switch (dbStatus) {
        case 'processing': return 'verifying';
        case 'completed': return 'verified';
        case 'failed': return 'failed';
        default: return 'verifying';
    }
}

@customElement('fb-photo-gallery')
export class FBPhotoGallery extends FBElement {
    static override styles = [
        FBElement.styles,
        css`
            :host {
                display: block;
            }

            .gallery {
                display: grid;
                grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
                gap: 12px;
            }

            .gallery-item {
                position: relative;
                border-radius: 8px;
                overflow: hidden;
                background: var(--fb-bg-tertiary, #1a1a1a);
                aspect-ratio: 1;
                cursor: pointer;
                transition: transform 0.15s ease;
            }

            .gallery-item:hover {
                transform: scale(1.02);
            }

            .gallery-item img {
                width: 100%;
                height: 100%;
                object-fit: cover;
                display: block;
            }

            .gallery-item__name {
                position: absolute;
                bottom: 0;
                left: 0;
                right: 0;
                padding: 4px 8px;
                background: linear-gradient(transparent, rgba(0,0,0,0.7));
                color: white;
                font-size: 11px;
                white-space: nowrap;
                overflow: hidden;
                text-overflow: ellipsis;
            }

            .empty {
                text-align: center;
                padding: 32px;
                color: var(--fb-text-secondary, #aaa);
                font-size: 14px;
            }
        `,
    ];

    /** List of assets to display */
    @property({ type: Array }) assets: GalleryAsset[] = [];

    /** Static utility re-exported for parent components that build GalleryAsset[] from API data */
    static mapAnalysisStatus = mapAnalysisStatus;

    private _handleAssetClick(asset: GalleryAsset): void {
        this.emit('fb-asset-click', { asset });
    }

    override render(): TemplateResult {
        if (this.assets.length === 0) {
            return html`<div class="empty">No photos uploaded yet</div>`;
        }

        return html`
            <div class="gallery">
                ${this.assets.map(asset => html`
                    <div
                        class="gallery-item"
                        @click=${() => { this._handleAssetClick(asset); }}
                    >
                        <img
                            src="${asset.file_url}"
                            alt="${asset.file_name}"
                            loading="lazy"
                        />
                        <fb-vision-badge
                            overlay
                            .status=${asset.analysis_status}
                            .summary=${asset.analysis_summary ?? ''}
                        ></fb-vision-badge>
                        <span class="gallery-item__name">${asset.file_name}</span>
                    </div>
                `)}
            </div>
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'fb-photo-gallery': FBPhotoGallery;
    }
}
