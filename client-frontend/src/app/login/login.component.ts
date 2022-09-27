import { Component, OnInit, Renderer2 } from '@angular/core';
import { Router } from '@angular/router';
import { environment } from 'src/environments/environment';
import { ScriptService } from '../core/services/script.service';

@Component({
  selector: 'app-login',
  templateUrl: './login.component.html',
  styleUrls: ['./login.component.css'],
  providers: [ ScriptService ]
})
export class LoginComponent implements OnInit {

  public hankoElementUrl = environment.hankoElementUrl;
  public hankoApiUrl = environment.hankoApiUrl;

  public isHankoElementsLoaded = false;

  constructor(private renderer: Renderer2,
    private scriptService: ScriptService,
    private readonly router: Router) { }

  ngOnInit() {
    const scriptElement = this.scriptService.loadJsScript(this.renderer, `${environment.hankoElementUrl}/element.hanko-auth.js`);
    scriptElement.onload = () => {
      this.isHankoElementsLoaded = true;
    };
    scriptElement.onerror = () => {
      console.error('Error loading elements');
    }
  }

  public redirectToIndex(e : any) {
    this.router.navigate(['/home']);
  }

}
