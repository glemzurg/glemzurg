# Inventory?

This use case gives Managers visibility into the status of Inventory: available 
Titles, current stock levels, and reorder levels.

◇

actors:
    manager: 

scenarios:

    inventory:
        name: Inventory
        details: Manager queries the inventory.
        objects:
            - key: pat
              name: Pat
              style: name
              class_key: 01_order_fulfillment/manager
            - key: title
              class_key: 01_order_fulfillment/title
              multi: true
            - key: medium
              class_key: 01_order_fulfillment/medium
              multi: true

        steps:
            - description: "*[Medium for Title]\n*?"
              from_object_key: medium
              to_object_key: title
              attribute_key: 01_order_fulfillment/medium/isbn
            - description: "*[*] *?"
              from_object_key: title
              to_object_key: pat
              attribute_key: 01_order_fulfillment/title/title
