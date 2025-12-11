# Drop from Cart

The Customer is removing a previously added Medium from their shopping cart. 
If there are no remaining Book Order Lines in this Book Order, this Book Order 
is also deleted.

◇

actors:
    customer: 

scenarios:

    add_to_cart:
        name: Simple
        details: The basic flow for dropping from a cart.

        objects:
            - key: joe
              name: Joe
              style: name
              class_key: 01_order_fulfillment/customer
            - key: order
              name: 47
              style: id
              class_key: 01_order_fulfillment/book_order
            - key: line_n
              name: n
              style: id
              class_key: 01_order_fulfillment/book_order_line
            - key: line
              class_key: 01_order_fulfillment/book_order_line
              multi: true

        steps:
 
            - from_object_key: joe
              to_object_key: order
              event_key: "01_order_fulfillment/book_order/event/drop"

            - from_object_key: order
              to_object_key: line_n
              event_key: "01_order_fulfillment/book_order_line/event/«destroy»"

            - from_object_key: line_n
              is_delete: true

            - description: "*[remaining Line in Book Order] *?"
              from_object_key: line
              to_object_key: order
              attribute_key: "01_order_fulfillment/book_order_line/quantity"

            - cases:
                - condition: No remaining Book Order Lines
                  statements:
                    - from_object_key: order
                      is_delete: true

