import {Component} from '@angular/core';
import {Http} from '@angular/http';
import {Observable} from 'rxjs/Observable';
import 'rxjs/add/operator/map'
import 'rxjs/add/operator/catch'

@Component({
  selector: 'app-root',
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.scss']
})
export class AppComponent {
  version: string;
  releaseVersion: string;
  updateAvailable: boolean;
  currentTab = 'marketplace';

  constructor(
    private http: Http,
  ) {
    // TODO(therealssj): set the version from somewhere
    this.version = "0.0.1";
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
}
