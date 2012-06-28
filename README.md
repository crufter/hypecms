HypeCMS
=======
A CMS written in Go, using MongoDB as a backend. (Work in progress, heavily.)

Installation
=======
- A MongoDB server running on the default port 27017. HypeCMS can start with an empty database, it will create everything it needs.
If you are running mongo on a different machine or port, modify the values in main.go.
- Install the next modules into your go environment:
	github.com/opesun/hypecms
	launchpad.net/mgo
	github.com/opesun/extract
	github.com/opesun/jsonp
	github.com/opesun/require
	github.com/opesun/routep
(Or anything else it whines for ;)

Demo
=======
Soon (in a week) you can test drive your own instance at hypecms.com (will be highly experimental though).

Design goals
=======
We hope we can write a system which can be used by Go based startups as a starting point/framework for development.
We try to build a system which is more oriented toward complete and unique web applications, rather than blogs or bussiness card web pages.
However at the same time, we try to keep the inner workings of it as simple as possible.
The CMS itself is just a thin layer above specific modules.
Everything can be overwritten, but at the same time there is a builtin default functionality provided.
Also, while performance is only a priority if it does no get in the way of clean code and maintainability, we believe that the Go and MongoDB is a killer combination.

Ideas and inner workings
=======
Any setup ever done to a site resides in the "Options" collection, the one with the latest date being the currently used option document.
The system handles option documents as immutable values. This allows easy backup and restoration of configuration. (You can switch back to any previous state, so there is no danger in installing or configuring plugins.)
Anything a site does, must be explicitly stated in this option document, with only two exceptions:

/admin
	- The admin module is hardcoded to avoid bootstrapping issues, eg. when a fresh site is born, we can issue basic commands and install modules trough it, even with an empty option document.
/debug
	- The debug module is hardcoded, so we can analyize the state of the site even with an empty or messed up option document.

The architecture is lousely based on the MVC pattern. Basically whenever the system receives a http request, it routes the given request according to the next rule:
- If the path of the request matches "/b/{modulename}", it will be routed to the given module, and the system expects a background operation to be called.
- Anything else will be handled as a "View" which will be displayed by either the builtin Display module or an explicitly stated one.

Some of the following things are awaiting implementation
-------

### The admin module
There is a builtin editor of the currently used option document, so you can configure any option or package by hand even if it has no graphical setup.

### The Display module
Every view function specifies a "Display Point". This display point contains information about what template file to execute (it contains this information implicitly, since the filename is the name of the display point itself).
Also, it can contain queries to be run. The queries are stored in the option document.
If the display module senses the Get parameter "json", it will not output the html as response, only a JSON encoded string containing the map[string]interface{} uni.Dat, which is the context of the template execution. (Not implemented yet.)
(Similarly, a background operation can receive "json" as a parameter, and instead of redirecting the user back to a view, it will print out the JSON encoded result of that background operation. Useful if we want ot build AJAX-based web applications.)

### The Content module (soon to come)
Every content in this module has a type, with user defined fields. For example, you can define a "blogpost" type and a "product" type, and install a webshop module to the "product" content type only.
Every type can have different fields and modules installed under it.

### The universe
The only value passed to modules is the *context.Uni (see api/context package)
For more reusable code, try to avoid doing bussiness logic in functions which has access to this value.

### "api/mod" package
Since there is no dynamic code loading in Go (yet), we must require all packages here.
Later if there will be a great number of modules we can compile with only the required modules.

(More things to come here...)

History
=======
HypeCMS is a port of an unreleased CMS written in Node.js.
The Node.js version has about 15 modules with 20 KLOC, we expect to port that in 3 months to the Go version, and even sooner if we get community support.

We don't intend to bash Node.js here, but we choose to continue this project in Go for the following reasons:
- In Go, there are no callbacks. When the codebase started to get bigger, we found async database access becoming a chore.
- In Node.js exceptions happening in async code is hard to catch. (Work being done on that, though.)
- We love the type system of Go. Although it's not as convenient to handle JSON-like objects, we find that the compiler aids implementation of a given idea.
- As a whole, we think Go is a much better designed language than JavaScript. Not that we don't enjoy writing JavaScript, we just enjoy writing Go more.
- Our line of thinking closely matches the Go author's. We value the beauty of simplicity, however hard is that to achieve.

Current status
=======
Work in progress.

License
=======
Released under the 2-clause BSD license, see license.txt file.