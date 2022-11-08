import { Injectable } from "@angular/core";
import { ServiceResponse } from "../models/service-response.interface";
import { BaseService } from "./service.base";

@Injectable()
export class PostService extends BaseService {
  public async getPosts(): Promise<ServiceResponse<GetPostsDto>> {
    return this.getAsync(`posts`);
  }

  public async createPost(post: CreatePostDto): Promise<ServiceResponse<void>> {
    return this.postAsync(`posts`, post);
  }
}

export interface GetPostsDto {
  posts: PostDto[];
}

export interface PostDto {
  id: string;
  created_at: Date;
  created_by: string;
  created_by_surrogate?: string;
  data: string;
}

export interface CreatePostDto {
  body: string;
}
