# A Basic Class

Here we have markdown details.

And even more.

◆

Here we have a UML comment.

And even more.

◇

actor_key: actor_key
attributes:
    - key: iso
      name: ISO Code
      details: The ISO 4217 code for the currency.
      rules: ref of valid ISO 4217 codes
      type_spec: STRING
      nullable: true
      invariants:
        - details: When an ISO code is set, it must be a valid ISO 4217 code.
          specification: "IF self.iso = NULL THEN TRUE ELSE self.iso \\in _Iso4217Codes"
    - key: country_code
      name: Country Code
      rules: ref of ISO 3166-1 two-letter codes
      type_spec: STRING
      nullable: true
      invariants:
        - details: When a country code is set, the country and optional state codes must match an allowed jurisdiction pair.
          specification: "IF self.country_code = NULL THEN TRUE ELSE IF self.state_code = NULL \\/ self.state_code = \"\" THEN <<self.country_code, \"\">> \\in _JurisdictionCodes ELSE <<self.country_code, self.state_code>> \\in _JurisdictionCodes"