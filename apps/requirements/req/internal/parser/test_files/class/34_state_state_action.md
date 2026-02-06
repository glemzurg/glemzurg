# A Basic Class

Here we have markdown details.

And even more.

◆

Here we have a UML comment.

And even more.

◇

actor_key: actor_key
states:
    Started:
        details: Appears in data dictionary.
        uml_comment: very import to users
        actions:
            - action: ProcessLog
              when: entry
actions:
    ProcessLog:
        details: Appears in data dictionary.
        requires:
            - userid (with details)
            - username (e.g. bob)
        guarantees:
            - userid (with details)
            - username (e.g. bob)
