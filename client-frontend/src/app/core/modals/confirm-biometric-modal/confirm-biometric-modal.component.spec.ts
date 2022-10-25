/* tslint:disable:no-unused-variable */
import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { MAT_DIALOG_DATA } from '@angular/material/dialog';
import { CommonTestingModule } from '../../../../testing/utils/CommonTestingModule';
import { ChallengeService } from '../../services/challenge.service';

import { ConfirmBiometricModalComponent } from './confirm-biometric-modal.component';

describe('ConfirmBiometricModalComponent', () => {
  let component: ConfirmBiometricModalComponent;
  let fixture: ComponentFixture<ConfirmBiometricModalComponent>;

  CommonTestingModule.setUpTestBed(ConfirmBiometricModalComponent, {});

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [ ConfirmBiometricModalComponent ],
      providers: [ { provide: MAT_DIALOG_DATA, useValue: "3280a1a2-9417-4b10-a6e9-987eabdf63ec" }, ChallengeService ]
    })
    .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(ConfirmBiometricModalComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
