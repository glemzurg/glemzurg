# A Basic Class

Here we have markdown details.

And even more.

◆

Here we have a UML comment.

And even more.

◇

actor_key: actor_key
attributes:
    - key: link_code
      name: Link Code
      rules: unconstrained
    - key: slot_num
      name: Slot Number
      rules: unconstrained
associations:
    - name: Contains
      details: Appears in data dictionary.
      from_multiplicity: "1"
      to_class_key: child_key
      to_multiplicity: any
      uniqueness:
        from_attributes:
            - link_code
            - slot_num
        to_attributes:
            - child_role
            - sort_order
      uml_comment: very import to users