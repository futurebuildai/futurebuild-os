-- Phase 7 Step 42: PostgreSQL LISTEN/NOTIFY trigger for feed_cards table.
-- Sends notifications on INSERT/UPDATE/DELETE so the SSE endpoint can push
-- real-time feed card events to connected frontend clients.
-- See FRONTEND_V2_SPEC.md §6.5

CREATE OR REPLACE FUNCTION notify_feed_changes() RETURNS trigger AS $$
DECLARE
    payload TEXT;
    card_id UUID;
    org_id UUID;
    op TEXT;
BEGIN
    op := TG_OP;
    IF TG_OP = 'DELETE' THEN
        card_id := OLD.id;
        org_id  := OLD.org_id;
    ELSE
        card_id := NEW.id;
        org_id  := NEW.org_id;
    END IF;

    payload := json_build_object(
        'op', op,
        'org_id', org_id,
        'card_id', card_id
    )::text;

    PERFORM pg_notify('feed_changes', payload);
    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER feed_cards_notify
    AFTER INSERT OR UPDATE OR DELETE ON feed_cards
    FOR EACH ROW EXECUTE FUNCTION notify_feed_changes();
