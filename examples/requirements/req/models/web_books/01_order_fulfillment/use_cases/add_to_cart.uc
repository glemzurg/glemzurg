# Add to Cart

In this use case, the Customer is adding one to many copies 
of the same Medium to their shopping cart. If no open 
Book Order exists for this Customer, a new Book Order is created.

◇

actors:
    customer: 

scenarios:

    add_to_cart:
        name: Simple
        details: The basic flow for adding to a cart.

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
              name: 1
              style: id
              class_key: 01_order_fulfillment/book_order_line
            - key: medium
              class_key: 01_order_fulfillment/medium


        steps:

            - description: "«new»(addressx)"
              from_object_key: joe
              to_object_key: order
              event_key: "01_order_fulfillment/book_order/event/«new»"

            - cases:
                - condition: No open Book Order for Customer
                  statements:
                    - description: "«new»(addressx)"
                      from_object_key: joe
                      to_object_key: order
                      event_key: "01_order_fulfillment/book_order/event/«new»"
 
            - description: "add(Medium,qtyx)"
              from_object_key: joe
              to_object_key: order
              event_key: "01_order_fulfillment/book_order/event/add"

            - description: "«new»(ord, Medium, qtyx)"
              from_object_key: order
              to_object_key: line
              event_key: "01_order_fulfillment/book_order_line/event/«new»"

            - description: "selling price"
              from_object_key: medium
              to_object_key: line
              attribute_key: "01_order_fulfillment/medium/cost"