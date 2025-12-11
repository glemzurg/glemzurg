# Packable Orders?

Warehouse Workers need visibility into which Book ORders (those involving Print Media) are packable.

◇

actors:
    warehouse_worker: 

scenarios:

    packable_orders:
        name: Simple
        details: All orders that are ready to be packed.
        objects:
            - key: al
              name: Al
              style: name
              class_key: 01_order_fulfillment/warehouse_worker
            - key: order
              class_key: 01_order_fulfillment/book_order
              multi: true
            - key: line
              class_key: 01_order_fulfillment/book_order_line
              multi: true
            - key: medium
              class_key: 01_order_fulfillment/medium_print

        steps:
            - description: "stock level?"
              from_object_key: medium
              to_object_key: line
              attribute_key: 01_order_fulfillment/medium_print/stock_level
            - description: "*[Line for Order]\nis packable?"
              from_object_key: line
              to_object_key: order
              attribute_key: 01_order_fulfillment/book_order_line/is_packable
            - description: "*[all Orders]\nis packable?"
              from_object_key: order
              to_object_key: al
              attribute_key: 01_order_fulfillment/book_order/is_packable