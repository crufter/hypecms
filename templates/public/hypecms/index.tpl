<!DOCTYPE html>

<html>
<head>
	<link rel="stylesheet" type="text/css" href="/template/style.css" />
	<script src="/shared/jquery.min.1.7.js"></script>
	<script src="/shared/anomal-RainbowVis-JS/rainbowvis.js"></script>
	<title>hypeCMS</title>
	<script>
		$(function(){
			var get = {};
			document.location.search.replace(/\??(?:([^=]+)=([^&]*)&?)/g, function () {
				function decode(s) {
					return decodeURIComponent(s.split("+").join(" "));
				}
				get[decode(arguments[1])] = decode(arguments[2]);
			});
			if (get["error"] != undefined) {
				$("#error").html(get["error"]).show()
			}
			if (get["ok"] != undefined) {
				if (get["-sitename"] == undefined) {
					$("#location").html("under yoursitename.hypecms.com")
				} else {
					$("#loc-link").attr("href", "http://"+get["-sitename"]+".hypecms.com")
				}
				$("#success").show()
			}
			var to = Math.floor(parseFloat($("#bar").attr("data-perc")))
			var bar_width = $("#bar").width()
			var span_width = Math.floor(bar_width/100)-1
			var numberOfItems = 100;
			var rainbow = new Rainbow(); 
			rainbow.setNumberRange(1, numberOfItems);
			rainbow.setSpectrum('green', 'red');
			for (i=0;i<to;i++) {
				var hexColour = rainbow.colourAt(i)
				$("#bar").append('<span style="float:left; display:block; width:'+span_width+'px; background:#'+hexColour+'; margin-right: 1px; solid;">&nbsp;</span>')
			}
			$("#bar").append('<div class="clearfix"></div>')
		})
	</script>
</head>
<body>
	<div id="fork_us_on_github"><a href="https://github.com/opesun/hypecms" TARGET="_blank"><img src="/template/fork-us-on-github.png"></a></div>
	<div id="gopher"><a href="" TARGET="_blank"><img src="/template/talks.png"></a></div>
	<div id="outer">
		<div id="inner">
			<h3 id="logo"><a href="https://github.com/opesun/hypecms" target="_blank">hypeCMS</a></h3>
			<br />
			<div id="under-logo">
				Test server load:<br /><br />
				<div id="perc">{{format_float .capacity_percentage 1}}%</div>
				<div id="bar" data-perc="{{.capacity_percentage}}">&nbsp;</div>
				<div class="clearfix"></div>
			</div>
			<div id="success">
				Everything went fine. In 5-10 seconds your site will be available <span id="location"><a id="loc-link" href="">here.</a></span>
				You can log in with the username "admin", and the password you provided.
			</div>
			<div id="error">
				&nbsp;
			</div>
		</div>
		<div id="login-box">
			<form action="/b/bootstrap/ignite" method="post">
				<span class="input-title">sitename:</span><br /><input title="yoo" type="text" title="Name of your site." name="sitename"><br /><br />
				<span class="input-title">admin password:</span><br /><input type="password" title="Administrator password you will use to log in to your site." name="password"><br /><br />
				<span class="input-title">password again:</span><br /><input type="password" title="Password confirmation." name="password_again"><br /><br />
				{{$p := show_puzzles "bootstrap" "ignite"}}
				{{html $p}}
				<input type="submit" value="Create new site">
			</form>
		</div>
		<div class="clearfix"></div>
		<div id="content">
			<h1>What is this about?</h1>
			<span class="col">
				HypeCMS is a fresh take on web application development. This 100% open source project aims to find the simplest way to be as expressive as possible.
				Grew out of our frustration with other cmses and frameworks, Hype is stripped down. For comparison, when configured as a WordPress-like blog (as in this demo), the front
				page costs you only 3 database queries at max (1 for the logged in user, 1 for site config,
			</span>
			<span class="col">
				1 for the contents), or even only 1 (0 query for not logged in users/if the users are turned off, 0 for site config because it is cached after first page load, 1 for queries).
				Aligning well with the no-bullshit approach of <a href="http://golang.org/">Go</a>, we intentionally design it in a way that you don't have to
				cut trough the fat. We hope that despite the fact that at the time of writing this there are no similar projects in Go, other developers share our view of it being an
			</span>
			<span class="col">
				excellent language for building web based applications, and we can join our efforts.
				We choose to use <a href="http://www.mongodb.org">mongoDB</a> as the database backend, because of it's flexibility of handling arbitrarily structured JSON data.
				Also, it's ability to scale ain't too shabby, which comes in handy when a lot of people start to use your shiny new application.
			</span>
			<div class="clearfix"></div>
			
			<br />
			<h2>Why the name?</h2>
			<p>
				The name stems of the common desire that most people want her web application to be hyped.<br />
				Hopefully, if you use hypeCMS, hype comes.
			</p>
			<br />
			<h2>What's working already?</h2>
			<p>
				For features, issues and source code, please refer to <a href="https://github.com/opesun/hypecms">github</a>.<br />
				This page is here only to provide provide a demo platform for possible early adopters and developers.<br />
				We realize that an open source project is only as good as it's documentation. We can't present it currently,<br />
				but we are actively working on it. Meanwhile, you can reach the primary developer under the nickname "wuttf" in the <a href="http://webchat.freenode.net/?channels=#go-nuts">#go-nuts IRC channel</a>, <br />
				or alternatively, feel free to post any issue on <a href="https://github.com/opesun/hypecms/issues/new">github</a>, even if it's a personal one.<br />
				You can also connect us by email, our address is "info", followed by a "@" symbol, and finally the domain opesun dot com.
			</p>
			<br />
			<h2>What are the plans for the future?</h2>
			<p>
				We don't have ten thousand 3rd party plugins (yet :P). To get around this, we wan't to include something like a "standard library",<br />
				which will be
				a collection of officially supported, hopefully high quality modules, which cover a large percentage of needs. Much like the standard library of a programming language.<br />
			</p>
			<br />
			<h2>How can I help the project?</h2>
			<p>
				Contribute :). Also, we would appreciate constructive criticism regarding design and code quality.<br />
				Hopefully smart people will chime in and help us improve the codebase.<br />
			</p>
			<h2>What's up with this test server?</h2>
			<p>
				We will try to keep this server alive, but don't host your most sensitive data here yet.<br />
			</p>
		</div>
		<div id="footer">
			<h4 style="font-size:90%;">Copyright 2012, the <a href="https://github.com/opesun/hypecms/blob/master/AUTHORS">hypeCMS authors.</a><h4>
		</div>
	</div>
</body>
</html>