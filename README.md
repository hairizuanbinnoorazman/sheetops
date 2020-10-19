# sheetops

Sheetops for demo

# Issues

## Unable to contact Google endpoint

The controller initially complains of the following error: `x509: certificate signed by unknown authority`. This is probably cause certificates are generally not shipped with the container. We would need to manually install it to get this portion working

## RBAC issues

One of the issues that came up on initial tests is where controller keeps complaining that it is unable to list deployments (which we kind of need here). Apparently, issue was we use the word `deployment` rather than `deployments`. Due to this typo, it is unable to pass the right permissions to service account, thereby causing issues.

## Missing permissions

Initial implementations of the controller have missing permissions for deployment object. It is initally assumed that the when creating resource, k8s only handles it implicitly; after creation, the resource does not matter to the controller. Unfortunately, this is not the case, apparently, once the object is created, the object is immediately under "watch" by the controller that created it.

Since initial implementation has the WATCH permission missing, the logs quickly got filled up with errors from missing permission issues, being unable to watch the newly created Deployment object

## Create Deployment object

The deployment object is not a simple object to create. There are various fields to fill up - when you run the controller, the controller panics out with one missing field at a time. The best way to kind of get a simple working example is to just see an example `deployment.yaml` and apply the same said fields - to quickly get something working

## Prevent reconcilation loops

Apparently, updating status on spec results in an update event being issues which triggers reconcilation. In the initial implementation of the controller, there are 3 status updates. There would result in 3 additional reconcialiation calls which would then call another 3 status updates. Essentially, this creates an infinite reconcilation loop where logs just stream by.

This causes additional issues as well. Since googlesheets api call is within this reconcilation loop, a sudden spike in api calls to google spreadsheets is made, resulting in error 429 from the library.

Issue is resolved by adding the predicate (a filter of sorts) - to check what kind of events will need to be processed by the controller

Refer to the following issue when trying to resolve this: https://github.com/operator-framework/operator-sdk/issues/2795
