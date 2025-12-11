# Cancel Order

The Customer has changed their mind about a Book Order after placing it and wants 
to cancel. The  Book Order will only be cancelled if it contains Print Media and 
has not already been packed.

◇

actors:
    customer: 

scenarios:

    cancel_order:
        name: Simple
        details: Customer discards their order.

        objects:
            - key: joe
              name: Joe
              style: name
              class_key: 01_order_fulfillment/customer
            - key: order
              name: 47
              style: id
              class_key: 01_order_fulfillment/book_order

        steps:

            - from_object_key: joe
              to_object_key: order
              event_key: "01_order_fulfillment/book_order/event/cancel"
