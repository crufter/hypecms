{{require header.t}}			
<div id="content-wrapper">
	<div class="container_16" id="content-wrapper2">
		<div class="grid_8" id="main-wrapper">
			<div class="main grid_8 section" id="main">
				<div class="widget Blog" id="Blog1">
				{{if .erorr}}
					{{.error}}
				{{else}} 						
					<div class="blog-posts hfeed">
					{{range .content_list}}
						<div class="post hentry uncustomized-post-template">
							<h3 class="post-title entry-title">
								<a href="/{{._id}}">{{.title}}</a>
							</h3>
							{{$tags := ._tags}}
							{{$user_name := ._users_created_by.name}}
							{{$created := .created}}
							{{require post_header.t}}
							<div class="post-body entry-content">
								<p>{{.excerpt}}</p>
								<div style="clear: both;"></div>
							</div>		
						{{require post_footer.t}}
						</div>
						{{end}}
					</div>	
					<div class="clear"></div>
				{{$navi := .content_list_navi}}
				<h3>{{require admin/navi.t}}</h3>
				{{end}}
				</div>
			</div>
		</div>	
		{{require sidebar.t}}
		<div class="clear"></div>
	</div>
</div>
{{require footer.t}}