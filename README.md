# httpext for Go

This package contains a set of utilities I've written (as well as one or two consolidated from other locations) that are useful when writting services that expose HTTP interfaces.

This package provides the following features:

+ a parser for the `Range` header;
+ a CORS header generator;
+ negotiation of content types per HTTP specification via the `Accept` header;
+ a standardized middleware interface;
+ a standardized, serializable way to provide representations of errors via HTTP.
