import { Component, OnInit } from '@angular/core';
import { NotificationService } from 'src/app/core/services/notification.service';
import { CreatePostDto, PostDto, PostService } from 'src/app/core/services/post.service';

@Component({
  selector: 'app-posts',
  templateUrl: './posts.component.html',
  styleUrls: ['./posts.component.css']
})
export class PostsComponent implements OnInit {

  public posts: PostDto[] = [];
  public postBody: string = "";

  constructor(private readonly postService: PostService, private readonly notificationService: NotificationService) { }

  async ngOnInit() {
    await this.getPosts();
  }

  async createPost() {
    console.log(this.postBody);
    if (!this.postBody || this.postBody.trim().length === 0 || this.postBody.trim().length > 150) {
      this.notificationService.error('post must be between 1 and 150 characters');
      return;
    }
    const dto: CreatePostDto = {
      body: this.postBody
    }
    const resp = await this.postService.createPost(dto);
    if (resp.type !== 'data') {
      this.notificationService.error('An error occurred making the post');
      return;
    }
    this.postBody = "";
    await this.getPosts();
  }

  async getPosts() {
    const postResp = await this.postService.getPosts();
    if (postResp.type !== 'data') {
      this.notificationService.error('An error occurred fetching posts!');
      return;
    }
    this.posts = postResp.data.posts;
  }

  public shouldShowSurrogate(post: PostDto): boolean {
    if (!post.created_by_surrogate) {
      return false;
    }
    return post.created_by_surrogate != post.created_by;
  }

}
