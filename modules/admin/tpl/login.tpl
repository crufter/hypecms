{{require admin/outer-header.t}}
<script src="/shared/jquery.min.1.7.js"></script>
<script src="/shared/chromahash/jquery.chroma-hash.js"></script>
<script>
$(function(){
$("input:password").chromaHash({bars: 3, salt:"7be82b35cb0199120eea35a4507c9acf", minimum:3});
})
</script>

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