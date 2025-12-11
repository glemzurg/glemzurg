# Check Out

The Customer has decided to proceed to checkout (i.e. place their Book Order).

◇

actors:
    customer: 

scenarios:

    check_out:
        name: Simple
        details: Purchasing the book order.

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

        steps:
 
            - from_object_key: joe
              to_object_key: order
              event_key: "01_order_fulfillment/book_order/event/place"

            - description: "*[Line in Order]\n"
              from_object_key: order
              to_object_key: line
              event_key: "01_order_fulfillment/book_order_line/event/place"
