# eBook User Keys?

The Customer is given user keys for any eBooks they have in a placed, 
packed, shipped, or completed Book Order.

◇

actors:
    customer: 

scenarios:

    ebook_user_keys:
        name: Simple
        details: All the eBook keys the user has.
        objects:
            - key: joe
              name: Joe
              style: name
              class_key: 01_order_fulfillment/customer
            - key: order
              name: 47
              style: id
              class_key: 01_order_fulfillment/book_order
            - key: line
              class_key: 01_order_fulfillment/book_order_line
              multi: true
            - key: medium
              class_key: 01_order_fulfillment/medium_ebook
            - key: title
              class_key: 01_order_fulfillment/title
            - key: publisher
              class_key: 01_order_fulfillment/publisher

        steps:

            - cases:
                - condition: Book Order has at least been placed, and has eBook Media
                  statements:

                    - loop: "eBook Medium for Book Order Line"
                      statements:

                        - description: "publisher key?"
                          from_object_key: publisher
                          to_object_key: title
                          attribute_key: 01_order_fulfillment/title/title
                        - description: "publisher key?"
                          from_object_key: title
                          to_object_key: medium
                          attribute_key: 01_order_fulfillment/title/title
                        - description: "user key?"
                          from_object_key: medium
                          to_object_key: line
                          attribute_key: 01_order_fulfillment/medium_ebook/user_key
                        - description: "user key?"
                          from_object_key: line
                          to_object_key: order
                          attribute_key: 01_order_fulfillment/medium_ebook/user_key

                    - description: "user keys?"
                      from_object_key: order
                      to_object_key: joe
                      attribute_key: 01_order_fulfillment/book_order/id


