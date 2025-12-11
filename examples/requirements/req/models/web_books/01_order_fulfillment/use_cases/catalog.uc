# Catalog?

This use case provides a Customer with the ability to browse the catalog:
visibility into Titles, Media, and their availability for ordering.

◇

actors:
    customer: 

scenarios:

    catalog:
        name: Simple
        details: A simple query of the available titles.
        objects:
            - key: joe
              name: Joe
              style: name
              class_key: 01_order_fulfillment/customer
            - key: title
              class_key: 01_order_fulfillment/title
              multi: true
            - key: medium
              class_key: 01_order_fulfillment/medium
              multi: true

        steps:
            - description: "*[Medium for Title and is visible]\nISBN, selling price?"
              from_object_key: medium
              to_object_key: title
              attribute_key: 01_order_fulfillment/medium/isbn
            - description: "[*customer interest]\ntitle, author, subject, ISBN(s), price(s)"
              from_object_key: title
              to_object_key: joe
              attribute_key: 01_order_fulfillment/title/title
