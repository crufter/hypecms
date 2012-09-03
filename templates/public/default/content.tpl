{{require header.t}}
<div id="content-wrapper">
	<div class="container_16" id="content-wrapper2">
		<div class="grid_8" id="main-wrapper">
			<div class="main grid_8 section" id="main">
				<div class="widget Blog" id="Blog1">
					<div class="blog-posts hfeed">
						<div class="post hentry uncustomized-post-template">
							<h3 class="post-title entry-title">{{.content.title}}</h3>
							{{$tags := .content._tags}}
							{{$user_name := .content._users_created_by.name}}
							{{$created := .content.created}}
							{{require post_header.t}}	
							<div class="post-body entry-content">
								<p>{{.content.content}}</p>
								{{require comment_listing.t}}
								{{require comment_insert.t}}
							</div>
						</div>
					</div>		
					<div class="clear"></div>
				</div>
			</div>
		</div>
		{{require sidebar.t}}
	</div>
</div>
{{require footer.t}}