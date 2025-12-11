# Order Status?

This use case gives the Customer visibility into the status of one of their Book 
Orders.

◇

actors:
    customer: 

scenarios:

    order_status:
        name: Simple
        details: Customer asks where this order is in its process.
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
            - description: "status?"
              from_object_key: order
              to_object_key: joe
              attribute_key: 01_order_fulfillment/book_order/id


