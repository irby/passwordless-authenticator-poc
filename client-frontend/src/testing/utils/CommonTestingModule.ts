import { HttpClientTestingModule } from "@angular/common/http/testing";
import { ChangeDetectionStrategy, ChangeDetectorRef } from "@angular/core";
import { TestBed, waitForAsync } from "@angular/core/testing";
import { FormsModule, ReactiveFormsModule } from "@angular/forms";
import {
    MatDialogModule,
    MAT_DIALOG_DATA,
    MatDialogRef,
} from "@angular/material/dialog";
import { RouterTestingModule } from "@angular/router/testing";
import { ToastrModule } from "ngx-toastr";
import { MatMenuModule } from "@angular/material/menu";
import { MatNativeDateModule, MatOptionModule } from "@angular/material/core";
import { MatInputModule } from "@angular/material/input";
import { MatSelectModule } from "@angular/material/select";
import { BrowserModule } from "@angular/platform-browser";
import {  NoopAnimationsModule } from "@angular/platform-browser/animations";
import { CUSTOM_ELEMENTS_SCHEMA } from "@angular/compiler";
import { MatError, MatFormField, MatFormFieldControl, MatFormFieldModule } from "@angular/material/form-field";
import { MatCheckboxModule } from "@angular/material/checkbox";
import { MatDatepickerModule } from "@angular/material/datepicker";
import { MatSortModule } from "@angular/material/sort";
import { MatTabsModule } from "@angular/material/tabs";
import { MatTableModule } from "@angular/material/table";
import { MatProgressBarModule } from "@angular/material/progress-bar";
import { MatProgressSpinnerModule } from "@angular/material/progress-spinner";
import { MatExpansionModule } from "@angular/material/expansion";
import { MatIconModule } from "@angular/material/icon";
import { MatTooltipModule } from "@angular/material/tooltip";
import { MatSlideToggleModule } from "@angular/material/slide-toggle";
import { AuthenticationService } from "../../app/core/services/authentication.service";
import { NotificationService } from "../../app/core/services/notification.service";
import { MockNotificationService } from "../mocks/mock.notification-service";

interface TestBedMetaData {
    providers?: object[],
    imports?: any[],
    declarations?: any[]
}

export class CommonTestingModule {
    public static setUpTestBed = (TestComponent: any, testBedMetaData?: TestBedMetaData) => {
        beforeEach(waitForAsync(() => {
            TestBed.configureTestingModule({
                imports: [
                    RouterTestingModule,
                    HttpClientTestingModule,
                    ToastrModule.forRoot(),
                    MatDialogModule,
                    FormsModule,
                    MatMenuModule,
                    MatCheckboxModule,
                    MatFormFieldModule,
                    MatInputModule,
                    MatSelectModule,
                    MatSortModule,
                    MatTabsModule,
                    MatTableModule,
                    NoopAnimationsModule,
                    MatDatepickerModule,
                    MatNativeDateModule,
                    MatProgressBarModule,
                    MatProgressSpinnerModule,
                    MatExpansionModule,
                    MatIconModule,
                    MatTooltipModule,
                    MatSlideToggleModule,
                    ...(testBedMetaData?.imports || [])
                ],
                declarations: [TestComponent, ...(testBedMetaData?.declarations || [])],
                providers: [
                    AuthenticationService,
                    { provide: MatDialogRef, useValue: {} },
                    {
                        provide: ChangeDetectorRef,
                        useValue: { detectChanges: jest.fn() },
                    },
                    { provide: NotificationService, useValue: MockNotificationService },
                    ...(testBedMetaData?.providers || []),
                ],
                schemas: [CUSTOM_ELEMENTS_SCHEMA],
            }).compileComponents();
        }));
    };
}
