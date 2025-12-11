# Send Replenish

The wait time (24 hours) has passed after a Replenish Order was opened, and the 
contents of the order need to be sent to the Publisher.

◇

actors:
    publisher: 

scenarios:

    send_replenish:
        name: Send Repenish
        details: Inform the Publisher that restocks are needed.
        objects:
            - key: line
              class_key: 01_order_fulfillment/replenish_order_line
              multi: true
            - key: order
              class_key: 01_order_fulfillment/replenish_order
            - key: publisher
              name: Permanent Press
              style: name
              class_key: 01_order_fulfillment/publisher

        steps:

            - description: "*[Line in Order]\n(title, qty)"
              from_object_key: line
              to_object_key: order
              attribute_key: "01_order_fulfillment/replenish_order_line/quantity"

            - from_object_key: order
              to_object_key: publisher
              event_key: "01_order_fulfillment/publisher/event/placeorder"

