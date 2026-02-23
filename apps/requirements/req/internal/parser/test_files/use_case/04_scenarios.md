# A Basic Use Case?

Here we have markdown details.

And even more.

◆

Here we have a UML comment.

And even more.

◇

level: sky

actors:
    actor_key: Actor stuff
    actor_b_key: Actor B stuff

scenarios:

    scenario_a_key:
        name: Scenario A
        details: Scenario A stuff.
        objects:
            - key: bob
              name: Bob
              style: name
              class_key: person_key
            - key: book
              name: 1
              style: id
              class_key: book_key
              multi: true
              uml_comment: the book of greatness
        steps:
            - key: "1"
              step_type: leaf
              leaf_type: event
              description: first step
              from_object_key: bob
              to_object_key: book
              event_key: class_key/event/processlog
            - key: "2"
              step_type: loop
              condition: while condition
              statements:
                  - key: "3"
                    step_type: leaf
                    leaf_type: scenario
                    description: loop step
                    from_object_key: book
                    to_object_key: bob
                    scenario_key: use_case_key/scenario/scenario_b_key
            - key: "4"
              step_type: switch
              statements:
                  - key: "5"
                    step_type: case
                    condition: case1
                    statements:
                        - key: "6"
                          step_type: leaf
                          leaf_type: event
                          description: case1 step
                          from_object_key: bob
                          to_object_key: book
                          event_key: class_key/event/processlog
                  - key: "7"
                    step_type: case
                    condition: case2
                    statements:
                        - key: "8"
                          step_type: leaf
                          leaf_type: delete
                          from_object_key: book

    scenario_b_key:
        name: Scenario B
        details: Scenario B stuff.
        objects:
            - key: bob
              name: Bob
              style: name
              class_key: person_key
            - key: book
              class_key: book_key