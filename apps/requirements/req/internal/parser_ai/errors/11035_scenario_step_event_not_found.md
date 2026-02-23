# ErrTreeScenarioStepEventNotFound (11035)

Description

- A scenario step references an `event_key` that does not exist on the resolved class's state machine.

Cause

- The step's `event_key` is incorrect or the class referenced by the step's object does not have that event defined in its `state_machine.events` map.

Why this matters

- Events are used to describe transitions and interactions. Missing event references make scenario steps ambiguous and prevent correct sequence or state reasoning.

How to fix

- Ensure the event exists in the referenced class's `state_machine.events` map.
- Fix typos in the `event_key` or correct the object-to-class mapping so the step resolves to the intended class.

Example

- Invalid: step references `event_key: confirm_order` but the class state machine defines only `confirm`.
- Fix: either change the event key to `confirm` or add the `confirm_order` event to the class state machine.
