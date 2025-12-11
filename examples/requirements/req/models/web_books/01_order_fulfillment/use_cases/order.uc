# Order?

This use case gives the Custoemr visibility into the contents of one of their Book Orders.

◇

actors:
    customer: 

scenarios:

    order:
        name: Simple
        details: Details about an order.
        objects:
            - key: joe
              name: Joe
              style: name
              class_key: 01_order_fulfillment/customer
            - key: order
              name: 47
              style: id
              class_key: 01_order_fulfillment/book_order
            - key: line
              class_key: 01_order_fulfillment/book_order_line
              multi: true
            - key: medium
              class_key: 01_order_fulfillment/medium
            - key: title
              class_key: 01_order_fulfillment/title

        steps:
            - description: "(title, author, subject)?"
              from_object_key: title
              to_object_key: medium
              attribute_key: 01_order_fulfillment/title/title
            - description: "(ISBN, title, author, subject)?"
              from_object_key: medium
              to_object_key: line
              attribute_key: 01_order_fulfillment/medium/isbn
            - description: "*[Line for Order]\n(ISBN, title, author, subject, price, qty)?"
              from_object_key: line
              to_object_key: order
              attribute_key: 01_order_fulfillment/book_order_line/quantity
            - description: "Order?"
              from_object_key: order
              to_object_key: joe
              attribute_key: 01_order_fulfillment/book_order/id


