import { Injectable } from '@angular/core';
import { Observable } from 'rxjs/Observable';
import { ErrorObservable } from 'rxjs/observable/ErrorObservable';
import { of } from 'rxjs/observable/of';
import { map, catchError } from 'rxjs/operators';
import { extend } from 'lodash';
import { BehaviorSubject } from 'rxjs/BehaviorSubject';

@Injectable()
export class ErrorScreenService {

  private errorSource = new BehaviorSubject<any>(false);
  currentError = this.errorSource.asObservable();

  constructor() { }

  setError(error: any) {
    this.errorSource.next(error)
  }

  clearError(){
    this.errorSource.next(false);
  }
}
