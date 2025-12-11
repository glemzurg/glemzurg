# Sales?

This use case gives Managers visibility into sales results over some period of time.

◇

actors:
    manager: 

scenarios:

    sales:
        name: Sales
        details: Manager queries the sales numbers.
        objects:
            - key: pat
              name: Pat
              style: name
              class_key: 01_order_fulfillment/manager
            - key: order
              class_key: 01_order_fulfillment/book_order
              multi: true
            - key: line
              class_key: 01_order_fulfillment/book_order_line
              multi: true

        steps:
            - description: "*[Lines for Order]\nsales?"
              from_object_key: line
              to_object_key: order
              attribute_key: 01_order_fulfillment/book_order_line/sales
            - description: "*[Book Order in date range]\nsales?"
              from_object_key: order
              to_object_key: pat
              attribute_key: 01_order_fulfillment/book_order/sales
