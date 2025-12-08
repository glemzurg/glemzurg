[⇦ WebBooks 2.0](model.md)

# Customer

This is an actor that represent persons or businesses that may place Book Orders with WebBooks.

## Classes

The domain classes that define this actor.



### Order Fulfillment : Customer

This class is an actor that represent persons or businesses that may place Book Orders with WebBooks. That person or business doesn't need to have placed any Book Orders to be considered a Customer; they only need to on WebBooks' marketing list. This class has no state diagram because it is an actor and is mostly external to Order Fullfillment.

| Name | Rules | Nullable | Comment |
| ---- | ----- | -------- | ------- |
| email address | any valid email address according to relevant Internet Engineering Task Force (IETF) specifications.   | false | The address that will be used for email communication between this customer and WebBooks. Customers may have more than one valid email address, but they mjust select only one of them for primary communication with WebBooks. |
| name | unconstrained   | false | The name of the custoemr--how they woudl like to be referred to--such as "John Smith" or "ABC Corp." |
| postal address | any mailing address acceptable to the US Post Office or Canada Post   | false | The primary psotal mailing address of the customer. A customer may h ave more than one valid postal addresss but must choose only one to be theior contact address. This is likely the customer's billing address. |
