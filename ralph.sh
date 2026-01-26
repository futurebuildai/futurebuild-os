#!/bin/bash
# The Ralph Wiggum Protocol

PLAN_FILE="LAUNCH_PLAN.md"
PROMISE="MISSION_COMPLETE"

while true; do
  # Run Claude with the prompt and auto-accept flag
  # We tell it to look at the plan and finish the next item.
  claude -p "Read $PLAN_FILE. Execute the next incomplete task. When the whole file is done, output '$PROMISE'." --auto-accept

  # Check if the last output contained the completion promise
  if grep -q "$PROMISE" "$PLAN_FILE" || [ $? -eq 0 ]; then
    echo "Ralph has finished the mission!"
    break
  fi

  echo "Task incomplete. Restarting Ralph for next iteration..."
done
