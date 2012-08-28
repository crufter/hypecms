<div class="navi">
	<link rel="stylesheet" type="text/css" href="/shared/scrollnavi/style.css" />
	<script src="/shared/scrollnavi/scrollnavi.js"></script>
	<script>
		$(function(){
			$("#navi_helper").scrollNavi({"cp":{{$navi.Current_page}},"results":{{$navi.All_results}},"rpp":10,"url":{{$navi.Url}}})
		})
	</script>
	{{range $navi.Result}}
		{{if .IsDot}}
			...
		{{else}}
			<a href="{{.Url}}">{{.Page}}</a> 
		{{end}}
	{{end}}
	<br /><br />
	<div id="navi_helper" style="width: 200px;"></div>
</div>