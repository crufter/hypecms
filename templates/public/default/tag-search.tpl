{{require header.t}}
<div id="content-wrapper">
	<div class="container_16" id="content-wrapper2">
		<div class="grid_8" id="main-wrapper">
			<div class="main grid_8 section" id="main">
				<div class="widget Blog" id="Blog1">
					<div class="blog-posts hfeed">
						<div class="post hentry uncustomized-post-template">
							{{if .error}}
								{{.error}}
							{{else}}
								<h3 class="post-title entry-title">Tags:</h3>
								<form id="tag-search">
									<input type="text" name="search">
									<input type="submit">
								</form>
								<div class="post-body entry-content">
									{{range .tag_list}}
										<ul style="float:left !important; width:30% !important; overflow:hidden !important";>
											<li><a href="/tag/{{.slug}}" title="{{.name}}">{{.name}}({{.count}})</a></li>
										</ul>
									{{end}}
								</div>
								<div class="clear"></div>
								{{$navi := .tag_list_navi}}
								<h3>{{require admin/navi.t}}</h3>
							{{end}}
						</div>
					</div>
				</div>
			</div>
		</div>
		{{require sidebar.t}}
	</div>
</div>
{{require footer.t}}