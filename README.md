HypeCMS
=======
A CMS written in Go, using MongoDB as a backend. (Work in progress, heavily.)

Why the name
=======
Most people want his blog or web app to be hyped, right?

Installation
=======
- A MongoDB server running on the default port 27017. HypeCMS can start with an empty database, it will create everything it needs.
If you are running mongo on a different machine or port, modify the values in main.go.
- Install the next modules into your go environment:
	* github.com/opesun/hypecms
	* github.com/opesun/hypecms-shared
	* labix.org/v2/mgo
	* github.com/opesun/extract
	* github.com/opesun/jsonp
	* github.com/opesun/require
	* github.com/opesun/routep  
	* anything else it whines for
- Create a symbolic link in your file system to point from github.com/opesun/hypecms/shared to github.com/opesun/hypecms-shared.
This needs to be done to separate the gazillion included JS libraries from the hypecms codebase. (And still able to serve those files.)

Demo
=======
Soon (in a week) you can test drive your own instance at hypecms.com (will be highly experimental though).

Design goals
=======
Hopefully Go based startups can use this as starting point/framework for development to save time.
Our focus is oriented toward complete and unique web applications, rather than blogs or bussiness card web pages.
We try to keep the inner workings of this as simple as possible, but there is a long way ahead.
Everything can be overwritten, but at the same time there is a builtin default functionality provided for convenience.
Performance, performance, performance. Only if it does not get in the way of readability and maintainability though.

Random info about the app so you can catch our drift
=======
Any setup ever done to a site resides in the "options" collection, the one with the latest date being the currently used option document.
The system handles option documents as immutable values. This allows easy backup and restoration of configuration. (You can switch back to any previous state, so there is no danger in installing or configuring plugins.)
Anything a site does, must be explicitly stated in this option document, with only two exceptions:

/admin  
	- The admin module is hardcoded to avoid bootstrapping issues, eg. when a fresh site is born, we can issue basic commands and install modules trough it, even with an empty option document.  
/debug  
	- The debug module is hardcoded, so we can analyize the state of the site even with an empty or messed up option document.  

The architecture is lousely based on the MVC pattern. Basically whenever the system receives a http request, it routes the given request according to the next rule:
- If the path of the request matches "/b/{modulename}", it will be routed to the given module, and the system expects a background operation to be called.
- Anything else will be handled as a "view" which will be displayed by either the builtin Display module or an explicitly stated one.

### The admin module
There is a builtin editor of the currently used option document, so you can configure any option or package by hand even if it has no graphical setup.

### The Display module
Every view function specifies a "display point". This display point contains information about what template file to execute (it contains this information implicitly, since the filename is the name of the display point itself).
Also, it can contain queries to be run. The queries are stored in the option document.
If the display module senses the get parameter "json", it will not output the html as response, only a JSON encoded string containing the map[string]interface{} uni.Dat, which is the context of the template execution. (Not implemented yet.)
(Similarly, a background operation can receive "json" as a parameter, and instead of redirecting the user back to a view, it will print out the JSON encoded result of that background operation. Useful if we want ot build AJAX-based web applications.)

### The Content module
Every content in this module has a type, with user defined fields. For example, you can define a "blogpost" type and a "product" type, and install a webshop module to the "product" content type only.
Every type can have different fields and modules installed under it.

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