[⇦ WebBooks 2.0](model.md)

# Publisher

This is an actor that represents external companies who publish and distribute books that WebBooks sells.

## Classes

The domain classes that define this actor.



### Order Fulfillment : Publisher

This is an actor that represents external companies who publish and distribute books that WebBooks sells.
WebBooks obtains both physical copies of Print media and publisher keys for eBooks from teh relevant Publisher.
While actor classes generally don't have state models, this is a relatively rare 
exception in that it does have behavior relevant to WebBooks processes. That behavior is shown in the 
state diagram, and the remainder of the behavior is outside the scope of Order Fulfillment.

| Name | Rules | Nullable | Comment |
| ---- | ----- | -------- | ------- |
| email address | any valid email address acoording to relevant IETF specifications   | false | The address that will be used for all (email) communication between this Publisher and WebBooks. Publishers may have more than one email address, but they must select only one of them for primary communication with WebBooks. |
| name | unconstrained   | false | The name of the Publisher--how they would like to be referred to--such as "Permanent Press" or "JKL Publishing House". |
| postal address | any mailing address acceptable to the US Post Office and Canada Post   | false | The primary postal mailing address of the publsiher. A publisher may have more than one valid postal mailing address but must choose only one to be their primary contact mailing address. |
