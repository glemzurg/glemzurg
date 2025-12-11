# Receive Replenish

A Replenish Order for Print Media was received from the Publisher; Print Media 
stock needs to be increased to reflect the number of copies received.

◇

actors:
    warehouse_worker: 

scenarios:

    receive_replenish:
        name: Receive Repenish
        details: |
            Order ready for shipping. GWTW is Gone With The Wind.
            This is the first part of the scenario. If stock goes
            (or already is) low on one or more Print Media, the 
            Replenish scenario triggers.
        objects:
            - key: al
              name: Al
              style: name
              class_key: 01_order_fulfillment/warehouse_worker
            - key: order
              name: 9
              style: id
              class_key: 01_order_fulfillment/replenish_order
            - key: line
              class_key: 01_order_fulfillment/replenish_order_line
              multi: true
            - key: medium
              class_key: 01_order_fulfillment/medium_print

        steps:

            - from_object_key: al
              to_object_key: order
              event_key: "01_order_fulfillment/replenish_order/event/received"

            - description: "*[Line in Order]\n"
              from_object_key: order
              to_object_key: line
              event_key: "01_order_fulfillment/replenish_order_line/event/received"

            - from_object_key: line
              to_object_key: medium
              event_key: "01_order_fulfillment/medium_print/event/restock"

            - from_object_key: line
              is_delete: true

            - from_object_key: order
              is_delete: true              