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
            - description: first step
              from_object_key: bob
              to_object_key: book
              event_key: class_key/event/processlog
            - loop: while condition
              statements:
                - description: loop step
                  from_object_key: book
                  to_object_key: bob
                  scenario_key: use_case_key/scenario/scenario_b_key
            - cases:
                - condition: case1
                  statements:
                    - description: case1 step
                      from_object_key: bob
                      to_object_key: book
                      event_key: class_key/event/processlog
                - condition: case2
                  statements:
                    - from_object_key: book
                      is_delete: true        

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

