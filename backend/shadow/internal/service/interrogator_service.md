# interrogator_service

## Intent
*   **High Level:** Onboarding agent that extracts project data from user conversations and documents
*   **Business Value:** Enables chat-first project creation by parsing natural language and blueprints

## Responsibility
*   Process onboarding messages (text and/or documents)
*   Extract structured project data using Gemini AI
*   Track conversation state and determine next questions
*   Detect long-lead procurement items from blueprints

## Key Logic
*   **ProcessMessage():** Main entry point - handles both document upload and text message branches
*   **extractFromDocument/extractFromBytes:** Uses Vision API to extract data from blueprints
*   **parseUserMessage:** Sends user message to Gemini with MessageParsingPrompt, strips markdown code blocks from response
*   **stripMarkdownCodeBlock:** Removes ```json ... ``` wrappers that Gemini adds to JSON responses
*   **getNextQuestion:** Checks priority fields (name, address, start_date, square_footage) to determine what to ask next
*   **checkReadyToCreate:** Returns true when name + address are present
*   **enrichLongLeadItems:** Adds estimated lead times based on known brand lead times

## Data Flow
1. Frontend sends message/document to `/api/v1/agent/onboard`
2. Service extracts values using AI (text or vision)
3. Merges extracted values with current state
4. Determines if ready to create or returns next question
5. Response includes extracted_values, confidence_scores, reply, ready_to_create

## Dependencies
*   **Upstream:** OnboardHandler (HTTP endpoint)
*   **Downstream:** ai.Client (Gemini), prompts package
