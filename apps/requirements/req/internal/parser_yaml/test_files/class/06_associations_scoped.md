# Scoped Association Keys

Cross-subdomain and cross-domain association targets use compact scoped keys.

◇

associations:
    - name: Uses Local
      from_multiplicity: "1"
      to_class_key: local_target
      to_multiplicity: "1"
    - name: Uses Cross Subdomain
      from_multiplicity: "1"
      to_class_key: other_subdomain/remote_class
      to_multiplicity: any
    - name: Uses Cross Domain
      from_multiplicity: "1"
      to_class_key: other_domain/other_subdomain/remote_class
      to_multiplicity: any