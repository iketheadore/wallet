import { Component, HostListener, ViewChild } from '@angular/core';
import { Http } from '@angular/http';
import { Observable } from 'rxjs/Observable';
import 'rxjs/add/operator/map'
import 'rxjs/add/operator/catch'
import { ErrorScreenService } from './error_screen/error_screen.service';
import { MatDialog } from '@angular/material';
import { SettingsComponent } from './settings/settings.component';
import { WalletAppModule } from 'wallet-lib';

@Component({
  selector: 'app-root',
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.scss']
})
export class AppComponent {
  version: string;
  releaseVersion: string;
  updateAvailable: boolean;
  currentTab = 'wallet';

  constructor(
    private http: Http,
    private errorScreenService: ErrorScreenService, 
    public dialog: MatDialog,
    private appMod: WalletAppModule
  ) {

    this.version = "0.0.0";

    if (window['require'])
    {
      this.version = window['require']('electron').remote.app.getVersion();
    }
    
    this.updateAvailable = false;
    this.retrieveReleaseVersion();
  }


  private higherVersion(first: string, second: string): boolean {
    const fa = first.split('.');
    const fb = second.split('.');
    for (let i = 0; i < 3; i++) {
      const na = Number(fa[i]);
      const nb = Number(fb[i]);
      if (na > nb || !isNaN(na) && isNaN(nb)) {
        return true;
      } else if (na < nb || isNaN(na) && !isNaN(nb)) {
        return false;
      }
    }
    return false;
  }

  private retrieveReleaseVersion() {
    this.http.get('https://api.github.com/repos/kittycash/wallet/tags')
      .map((res: any) => res.json())
      .catch((error: any) => Observable.throw(error || 'Unable to fetch latest release version from github.'))
      .subscribe(response => {
        let tagElem = response.find(element => element['name'].indexOf('rc') === -1);
        if (tagElem !== undefined) {
          this.releaseVersion = tagElem['name'].substr(1);
          this.updateAvailable = this.higherVersion(this.releaseVersion, this.version);
        }
      });
  }

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

  doOpenSettings(){
    this.dialog.open(SettingsComponent, { width: '700px' });
  }

  toggleBar() {
    let sidebar = document.getElementById("wallet_sidebar");
    if (sidebar.style.display == "none")
    {
      sidebar.style.display = "block";
    }
    else
    {
      sidebar.style.display = "none";
    }
  }
}