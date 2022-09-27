import { Component, OnInit } from '@angular/core';
import { Router } from '@angular/router';
import { AuthenticationService } from '../core/services/authentication.service';

@Component({
  selector: 'app-default',
  templateUrl: './default.component.html',
  styleUrls: ['./default.component.css']
})
export class DefaultComponent implements OnInit {

  constructor(private readonly router: Router) { }

  async ngOnInit() {
    if(await AuthenticationService.isAuthenticated()) {
      this.router.navigate(['/home'], { replaceUrl: true });
      return;
    }

    this.router.navigate(['/login'], { replaceUrl: true });
  }

}
