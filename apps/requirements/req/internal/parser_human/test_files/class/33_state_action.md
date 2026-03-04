# A Basic Class

Here we have markdown details.

And even more.

◆

Here we have a UML comment.

And even more.

◇

actor_key: actor_key
actions:
    ClearLog:
        details: Clears the log.
    ProcessLog:
        details: Appears in data dictionary.
        parameters:
            - name: Amount
              rules: Nat
            - name: Label
              rules: unconstrained
        requires:
            - details: The amount must be positive.
              specification: amount > 0
            - details: The label must not be empty.
        guarantees:
            - details: The log is updated.
              target: log
              specification: self.log' = Append(self.log, amount)
        safety_rules:
            - details: The total never exceeds the limit.
              specification: self.total' <= self.limit
