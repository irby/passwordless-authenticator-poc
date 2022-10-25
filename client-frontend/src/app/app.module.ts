import { CUSTOM_ELEMENTS_SCHEMA, NgModule } from '@angular/core';
import { BrowserModule } from '@angular/platform-browser';
import { MatDialogModule } from '@angular/material/dialog';
import { AppRoutingModule } from './app-routing.module';
import { AppComponent } from './app.component';
import { LoginComponent } from './login/login.component';
import { DefaultComponent } from './default/default.component';
import { HomeComponent } from './home/home.component';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { AccountSharingInitializationDialog } from './home/account-sharing-initialization/account-sharing-initialization.component';
import { MatCommonModule } from '@angular/material/core';
import { MatButtonModule } from '@angular/material/button';
import {MatTableModule} from '@angular/material/table';
import {MatIconModule} from '@angular/material/icon';
import {MatInputModule} from '@angular/material/input';
import {MatProgressSpinnerModule} from '@angular/material/progress-spinner';
import { FormsModule, ReactiveFormsModule } from '@angular/forms';
import {MatSlideToggleModule} from '@angular/material/slide-toggle';
import {MatProgressBarModule} from '@angular/material/progress-bar';
import { ShareComponent } from './share/share.component';
import { SocketService } from './core/services/socket.service';
import { UserService } from './core/services/user.service';
import { HttpClient, HttpClientModule } from '@angular/common/http';
import { AuthenticationService } from './core/services/authentication.service';
import { GrantService } from './core/services/grant.service';
import { GrantsGuestModalComponent } from './home/grants-guest-modal/grants-guest-modal.component';
import { GrantsParentModalComponent } from './home/grants-parent-modal/grants-parent-modal.component';
import { ChallengeService } from './core/services/challenge.service';
import { NotificationService } from './core/services/notification.service';
import { ToastrModule } from 'ngx-toastr';
import { ConfirmBiometricModalComponent } from './core/modals/confirm-biometric-modal/confirm-biometric-modal.component';

@NgModule({
  declarations: [				
    AppComponent,
      LoginComponent,
      DefaultComponent,
      AccountSharingInitializationDialog,
      HomeComponent,
      ShareComponent,
      GrantsGuestModalComponent,
      GrantsParentModalComponent,
      ConfirmBiometricModalComponent
   ],
  imports: [
    BrowserModule,
    MatCommonModule,
    AppRoutingModule,
    MatDialogModule,
    BrowserAnimationsModule,
    MatButtonModule,
    MatInputModule,
    MatSlideToggleModule,
    FormsModule,
    MatIconModule,
    ReactiveFormsModule,
    MatProgressBarModule,
    MatProgressSpinnerModule,
    MatTableModule,
    HttpClientModule,
    ToastrModule.forRoot(),
  ],
  providers: [SocketService, AuthenticationService, UserService, GrantService, ChallengeService],
  bootstrap: [AppComponent],
  schemas: [CUSTOM_ELEMENTS_SCHEMA]
})
export class AppModule { }
