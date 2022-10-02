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

@NgModule({
  declarations: [			
    AppComponent,
      LoginComponent,
      DefaultComponent,
      AccountSharingInitializationDialog,
      HomeComponent
   ],
  imports: [
    BrowserModule,
    MatCommonModule,
    AppRoutingModule,
    MatDialogModule,
    BrowserAnimationsModule
  ],
  providers: [],
  bootstrap: [AppComponent],
  schemas: [CUSTOM_ELEMENTS_SCHEMA]
})
export class AppModule { }
