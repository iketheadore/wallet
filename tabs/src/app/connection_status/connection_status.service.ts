import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs/Observable';
import { of } from 'rxjs/observable/of';
import { map, catchError } from 'rxjs/operators';
import { BehaviorSubject } from 'rxjs/BehaviorSubject';
import { environment } from '../../environments/environment';

const routes = {
  status: () => `${environment.walletUrl}/ping`,
};

@Injectable()
export class ConnectionStatusService {


  private statusSource = new BehaviorSubject<any>(false);
  status = this.statusSource.asObservable();

  constructor(private httpClient: HttpClient) { 
    this.checkStatus();
  }

  private checkStatus()
  {
    //Now get the status
    this.getStatus().subscribe((response: any) => {

      let status = {connected: response.success, timer: 0, retry_time: 30};

      clearInterval(status.timer);
      if (!status.connected)
      {
        let __this = this;
        //Set a timer
        status.retry_time = 30;
        status.timer = setInterval(() => { 
          status.retry_time = status.retry_time - 1;
          if (status.retry_time <= 0)
          {
            clearInterval(status.timer);
            __this.doReconnect();
            return;
          }
        }, 1000);
      }

      this.statusSource.next(status);
    });
  }

  doReconnect()
  {
    this.checkStatus();
  }

 
  private getStatus(): Observable<any> {
    return this.httpClient
      .get(routes.status())
      .pipe(
        map((body: any) => body),
        catchError(() => of('Error, could not load status :-('))
      );
  }
}
