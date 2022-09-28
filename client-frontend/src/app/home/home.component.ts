import { Component, OnInit } from '@angular/core';
import { Router } from '@angular/router';
import { AuthenticationService } from '../core/services/authentication.service';

@Component({
  selector: 'app-home',
  templateUrl: './home.component.html',
  styleUrls: ['./home.component.css']
})
export class HomeComponent implements OnInit {

  constructor(private readonly router: Router) { }

  async ngOnInit() {
    if (!(await AuthenticationService.isAuthenticated())) {
      this.router.navigate(['']);
    }
  }

  public async logout() {
    await AuthenticationService.logout();
    this.router.navigate(['']);
  }

}