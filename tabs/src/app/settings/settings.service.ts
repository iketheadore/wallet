import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs/Observable';
import { of } from 'rxjs/observable/of';
import { map, catchError } from 'rxjs/operators';
import { environment } from '../../environments/environment';

const routes = {
  new_wallet: (s: WalletContext) => `http://127.0.0.1:6148/v1/wallets/new`,
  list_wallets: () => `http://127.0.0.1:6148/v1/wallets/list`,
  get_wallet: () => `http://127.0.0.1:6148/v1/wallets/get`
};

export interface WalletContext {
  label: string;
  seed: string;
  aCount: number;
  encrypted: boolean;
  password: string;
}

@Injectable()
export class SettingsService {

  constructor(private httpClient: HttpClient) { }

  restoreSeed(context: WalletContext): Observable<object> {
    return this.httpClient
      .post(routes.new_wallet(context), this.getQueryString(context), this.getOptions())
      .pipe(
        map((body: any) => body),
        catchError(() => of('Error, could not restore seed :-('))
      );
  }

  getBackupFile(wallets): Promise<any> {
    return new Promise<any>((resolve, reject) => {
      let promises = [];

      for (let i = 0; i < wallets.length; i++)
      {
        let wallet = wallets[i];
        promises.push(this.getWalletDetails(wallet.label, null));
      }
      Promise.all(promises).then(wallets=> {
        resolve(wallets);
      }, 
      error => {
        alert(error);
      })
    });
  }

  getWalletList(): Promise<any> {
    return new Promise<any>((resolve, reject) => {
      this.httpClient
      .get(routes.list_wallets(), this.getOptions()).subscribe((wallets:any) => {

        if (wallets && wallets.wallets)
        {
          resolve(wallets.wallets);
        }
        else
        {
          reject("No wallets found");
        }
        
      });
    });
  }

  getWalletDetails(label: string, password: any): Promise<any> {
    return new Promise<any>((resolve, reject) => {
      this.httpClient
      .post(routes.get_wallet(), this.getQueryString({label: label, password: password}), this.getOptions()).subscribe(

        (wallet:any) => {
          if (wallet && wallet.entry_count && wallet.entry_count > 0)
          {
            resolve(wallet);
          }        
          else
          {
            reject("Wallet not found");
          }
      }, (err: any) => {
        //Don't have proper error codes, assuming locked wallet for now
        alert("Warning: The wallet: " + label + " is locked.  It will not be included in the backup file.  If you wish to backup wallet: " + label + ", please unlock the wallet and run the backup again.");
        resolve(false);
      }
    );
  })
  }

  private getQueryString(parameters:any = null) {
    if (!parameters) {
      return '';
    }

    return Object.keys(parameters).reduce((array,key) => {
      array.push(key + '=' + encodeURIComponent(parameters[key]));
      return array;
    }, []).join('&');
  }


  private getOptions() {
    const headers = {
      'Content-Type': 'application/x-www-form-urlencoded',
    };

    return { headers: headers };
  }
}
