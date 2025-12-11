# Ship Order

A Book Order (with Print Media) was previously packed and has been taken by a shipper 
for delivery to the Customer.

◇

actors:
    warehouse_worker: 

scenarios:

    shipl_order:
        name: Ship Order
        details: A shipper has picked up the order.

        objects:
            - key: al
              name: Al
              style: name
              class_key: 01_order_fulfillment/warehouse_worker
            - key: order
              name: 47
              style: id
              class_key: 01_order_fulfillment/book_order

        steps:

            - from_object_key: al
              to_object_key: order
              event_key: "01_order_fulfillment/book_order/event/shipped"