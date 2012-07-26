{{require admin/outer-header.t}}

<div>
	<h3>Login</h3>
</div>
<form method="post" action="/admin/b/adminlogin">
	Name<br />
	<input name="name"><br />
	<br />
	Password<br />
	<input name="password" type="password"><br />
	<br />
	<input type="submit">
</form>

{{require admin/outer-footer.t}}