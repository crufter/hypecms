<!DOCTYPE html>
<html>
<head>
	<script src="/shared/jquery.min.1.7.js"></script>
	<script src="/shared/textinputs_jquery.js"></script>
	<script src="/tpl/admin/terminal.js"></script>
	<link rel="stylesheet" type="text/css" href="/tpl/admin/terminal.css" />
</head>

<body>
	<div id="display">
		&nbsp;
	</div>
	
	<div id="left">
	{{if ._user.name}}
		<div id="logged-in-as" data-user="{{._user.name}}">{{._user.name}} > </div>
	{{else}}
		<div id="logged-in-as" data-user="{{._user.guest_name}}">{{._user.guest_name}} > </div>
	{{end}}
	</div>
	
	<div id="right">
		<form id="terminal" action="/b-terminal">
			<textarea spellcheck=false id="terminal-inp" class="terminal-inp"></textarea>
		</form>
	</div>
	<div class="clearfix"></div>
	<div id="whitespace">&nbsp;</div>
</body>
</html>