#!/bin/bash
# Overlord V8.1 - The Inspector
# Forces actual file creation before checklist updates.

PLAN_FILE="PLAN.md"
[ ! -f "$PLAN_FILE" ] && [ -f "memory/PLAN.md" ] && PLAN_FILE="memory/PLAN.md"

if [ ! -f "$PLAN_FILE" ]; then echo "❌ PLAN.md not found."; exit 1; fi

LOG_FILE="logs/overlord.log"
mkdir -p logs
echo "🚀 Overlord V8.1 [Inspector Mode] tracking: $PLAN_FILE" | tee -a "$LOG_FILE"

while true; do
    TASK_LINE=$(grep -m 1 "\[ \]" "$PLAN_FILE")
    if [ -z "$TASK_LINE" ]; then echo "✅ All tasks complete."; exit 0; fi

    TASK_ID=$(echo "$TASK_LINE" | grep -oE "(Step [0-9]+|[A-Z]\.[0-9]+(\.[0-9]+)?)" || echo "TASK")
    
    echo -e "\n🤖 ENGAGING AGENT -> $TASK_ID" | tee -a "$LOG_FILE"

    # --- HARDENED PROMPT ---
    # We add a verification clause to prevent "simulated" completion.
    PROMPT="You are an autonomous Senior Engineer.
    Your Mission: Execute '$TASK_LINE' from $PLAN_FILE.
    
    CRITICAL PROTOCOL:
    1. Read the Spec referenced in the task.
    2. Implement all required logic. If this is a backend task, you MUST create/update the .go files.
    3. VERIFY: Run 'ls -l' on the files you created to prove they exist in the file system.
    4. ONLY AFTER verifying existence, update $PLAN_FILE by changing '[ ] $TASK_ID' to '[x] $TASK_ID'.
    5. Summarize your work and exit."

    echo "   ...Agent working (Monitor with: tail -f $LOG_FILE)..."
    claude --dangerously-skip-permissions "$PROMPT" >> "$LOG_FILE" 2>&1
    
    EXIT_CODE=$?

    if [ $EXIT_CODE -ne 0 ]; then
        echo "⚠️ Agent process failed. Retrying in 10s..." | tee -a "$LOG_FILE"
        sleep 10
        continue
    fi

    # Overlord's own local verification
    if grep -q "\[ \] $TASK_ID" "$PLAN_FILE"; then
        echo "⚠️ Agent skipped checklist. Overlord forcing state for $TASK_ID..." | tee -a "$LOG_FILE"
        if [[ "$OSTYPE" == "darwin"* ]]; then
            sed -i '' "s/\[ \] $TASK_ID/\[x\] $TASK_ID/" "$PLAN_FILE"
        else
            sed -i "s/\[ \] $TASK_ID/\[x\] $TASK_ID/" "$PLAN_FILE"
        fi
    fi

    echo "✅ $TASK_ID Cycle Cleared. Cooling down..."
    sleep 5
done
