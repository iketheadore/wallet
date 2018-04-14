import { Component, HostListener } from '@angular/core';
import { ErrorScreenService } from './error_screen/error_screen.service';

@Component({
  selector: 'app-root',
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.scss']
})
export class AppComponent {
  currentTab : string = 'marketplace';

  constructor(private errorScreenService: ErrorScreenService) { }

  @HostListener('document:showGlobalError', ['$event'])
	  onError(ev:any) {
	  	ev.preventDefault();
	    // send the error to the error screen service
	    this.errorScreenService.setError(ev.detail.message);
	  }

  doRefresh() {
    let event = new CustomEvent('refreshButtonClick', { cancelable: true, detail: {} });
    document.dispatchEvent(event);
  }
}
