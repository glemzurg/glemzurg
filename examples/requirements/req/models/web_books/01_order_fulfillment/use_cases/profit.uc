# Profit?

This use case gives Managers visibility into profit results over some period of time.

◇

actors:
    manager: 

scenarios:

    profit:
        name: Profit
        details: Manager queries the profits.
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
            - key: medium
              class_key: 01_order_fulfillment/medium

        steps:
            - description: "acquisition cost?"
              from_object_key: medium
              to_object_key: line
              attribute_key: 01_order_fulfillment/medium/cost
            - description: "*[Lines for Order]\nprofit?"
              from_object_key: line
              to_object_key: order
              attribute_key: 01_order_fulfillment/book_order_line/profit
            - description: "*[Book Order in date range]\nprofit?"
              from_object_key: order
              to_object_key: pat
              attribute_key: 01_order_fulfillment/book_order/profit