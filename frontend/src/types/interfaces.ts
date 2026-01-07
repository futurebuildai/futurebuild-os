import { Contact, Forecast } from "./models";

/**
 * WeatherService defines the integration for the SWIM Model.
 * See API_AND_TYPES_SPEC.md Section 2.1
 */
export interface WeatherService {
    getForecast(lat: number, long: number): Promise<Forecast>;
}

/**
 * VisionService defines the Validation Protocol service.
 * See API_AND_TYPES_SPEC.md Section 2.2
 */
export interface VisionService {
    /**
     * VerifyTask returns (is_verified, confidence_score)
     */
    verifyTask(imageURL: string, taskDescription: string): Promise<[boolean, number]>;
}

/**
 * NotificationService defines the outbound communication service.
 * See API_AND_TYPES_SPEC.md Section 2.3
 */
export interface NotificationService {
    sendSMS(contactID: string, message: string): Promise<void>;
}

/**
 * DirectoryService defines contact and assignment lookups.
 * See API_AND_TYPES_SPEC.md Section 2.4
 */
export interface DirectoryService {
    getContactForPhase(phaseID: string): Promise<Contact>;
}
