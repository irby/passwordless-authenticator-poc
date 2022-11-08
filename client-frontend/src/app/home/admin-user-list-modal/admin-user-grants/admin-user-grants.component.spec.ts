/* tslint:disable:no-unused-variable */
import { async, ComponentFixture, TestBed } from "@angular/core/testing";
import { By } from "@angular/platform-browser";
import { DebugElement } from "@angular/core";

import { AdminUserGrantsComponent } from "./admin-user-grants.component";
import { MAT_DIALOG_DATA } from "@angular/material/dialog";
import { UserModalInfo } from "../../../core/models/user-modal-info.interface";
import { AdminService } from "../../../core/services/admin.service";
import { CommonTestingModule } from "../../../../testing/utils/CommonTestingModule";

describe("AdminUserGrantsComponent", () => {
  let component: AdminUserGrantsComponent;
  let fixture: ComponentFixture<AdminUserGrantsComponent>;

  const userInfo: UserModalInfo = {
    userId: "test",
    userEmail: "test@example.com",
  };

  CommonTestingModule.setUpTestBed(AdminUserGrantsComponent, {});

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [AdminUserGrantsComponent],
      providers: [
        { provide: MAT_DIALOG_DATA, useValue: userInfo },
        AdminService,
      ],
    }).compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(AdminUserGrantsComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it("should create", () => {
    expect(component).toBeTruthy();
  });
});
