# Pack Order

A Book Order involving Print Media has been packed by a Warehouse Worker;
stock needs to be decreased by the number of copies packed.

◇

actors:
    warehouse_worker: 
    publisher: 

scenarios:

    pack_order:
        name: Pack Order
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
              name: 47
              style: id
              class_key: 01_order_fulfillment/book_order
            - key: line
              class_key: 01_order_fulfillment/book_order_line
              multi: true
            - key: medium
              class_key: 01_order_fulfillment/medium_print
            - key: title
              name: GWTW
              style: name
              class_key: 01_order_fulfillment/title

        steps:

            - from_object_key: al
              to_object_key: order
              event_key: "01_order_fulfillment/book_order/event/packed"

            - description: "*[Line in Order]\n"
              from_object_key: order
              to_object_key: line
              event_key: "01_order_fulfillment/book_order_line/event/packed"

            - from_object_key: line
              to_object_key: medium
              event_key: "01_order_fulfillment/medium_print/event/picked"

            - cases:
                - condition: Stock low
                  statements:
                    - from_object_key: medium
                      to_object_key: title
                      scenario_key: "01_order_fulfillment/pack_order/scenario/replenish"


    replenish:
        name: Replenish
        details: |
            A Book Order involving Print Media got packed adn stock went 
            (or already was) low on one or more Print Media, triggering 
            the need to replenish stock. This is the second part of the 
            Pack Order sequence diagram.

        objects:
            - key: medium
              name: GWTW
              style: name
              class_key: 01_order_fulfillment/medium_print
            - key: title
              name: GWTW
              style: name
              class_key: 01_order_fulfillment/title
            - key: publisher
              name: Permanent Press
              style: name
              class_key: 01_order_fulfillment/publisher
            - key: order
              name: 9
              style: id
              class_key: 01_order_fulfillment/replenish_order
            - key: line
              name: n
              style: id
              class_key: 01_order_fulfillment/replenish_order_line

        steps:

            - from_object_key: medium
              to_object_key: title
              event_key: "01_order_fulfillment/title/event/replenish"

            - from_object_key: title
              to_object_key: publisher
              event_key: "01_order_fulfillment/publisher/event/replenish"

            - cases:
                - condition: No open Replenish Book Order
                  statements:
                    - from_object_key: publisher
                      to_object_key: order
                      event_key: "01_order_fulfillment/replenish_order/event/«new»"
 
            - from_object_key: publisher
              to_object_key: order
              event_key: "01_order_fulfillment/book_order/event/add"

            - cases:

                - condition: No Replenish Order Line for Medium
                  statements:
                    - from_object_key: order
                      to_object_key: line
                      event_key: "01_order_fulfillment/replenish_order/event/«new»"
            
                - condition: Replenish Order Line for Medium
                  statements:
                    - from_object_key: order
                      to_object_key: line
                      event_key: "01_order_fulfillment/replenish_order/event/add"


