# A Basic Class

Here we have markdown details.

And even more.

◆

Here we have a UML comment.

And even more.

◇

actor_key: actor_key
queries:
    GetBalance:
        details: Returns the current balance.
        parameters:
            - name: AccountId
              rules: Nat
        requires:
            - details: The account must exist.
              specification: accountId \in DOMAIN self.accounts
        guarantees:
            - details: Returns the balance for the account.
              target: balance
              specification: self.accounts[accountId].balance
    ListAccounts:
        details: Lists all accounts.
