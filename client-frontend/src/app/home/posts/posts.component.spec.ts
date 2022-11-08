/* tslint:disable:no-unused-variable */
import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { PostsComponent } from './posts.component';
import { CommonTestingModule } from '../../../testing/utils/CommonTestingModule';
import { PostDto, PostService } from '../../../app/core/services/post.service';
import { MockNotificationService } from '../../../testing/mocks/mock.notification-service';
import { NotificationService } from '../../../app/core/services/notification.service';
import axios from 'axios';

describe('PostsComponent', () => {
  let component: PostsComponent;
  let fixture: ComponentFixture<PostsComponent>;

  jest.mock("axios");
  axios.get = jest.fn().mockImplementation(() => {});

  CommonTestingModule.setUpTestBed(PostsComponent, {});

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [ PostsComponent ],
      providers: [ {provide: NotificationService, useValue: MockNotificationService}, PostService ]
    })
    .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(PostsComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });

  it('should return false if created_by_surrogate is null', () => {
    const post: PostDto = {
      id: "myid",
      created_at: new Date(),
      created_by: "hello@example.com",
      created_by_surrogate: undefined,
      data: "eeeee"
    }
    const result = component.shouldShowSurrogate(post);
    expect(result).toBe(false);
  })

  it('should return false if created_by_surrogate is not null but created_by is the same', () => {
    const post: PostDto = {
      id: "myid",
      created_at: new Date(),
      created_by: "hello@example.com",
      created_by_surrogate: "hello@example.com",
      data: "eeeee"
    }
    const result = component.shouldShowSurrogate(post);
    expect(result).toBe(false);
  })

  it('should return false if created_by_surrogate is not null and created_by is not the same', () => {
    const post: PostDto = {
      id: "myid",
      created_at: new Date(),
      created_by: "hello@example.com",
      created_by_surrogate: "world@example.com",
      data: "eeeee"
    }
    const result = component.shouldShowSurrogate(post);
    expect(result).toBe(true);
  })
});
