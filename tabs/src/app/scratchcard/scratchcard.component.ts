import { Component, OnInit, Directive, ElementRef, HostListener, HostBinding, Renderer2, Input, Output, EventEmitter } from '@angular/core';
import { finalize } from 'rxjs/operators';
import { HttpClient } from '@angular/common/http';
import { environment } from '../../environments/environment';
import { MatDialog } from '@angular/material';
import { FormGroup, FormBuilder, Validators} from '@angular/forms';
import { SettingsService } from '../settings/settings.service';

@Component({
  selector: 'scratchcard',
  templateUrl: './scratchcard.component.html',
  styleUrls: ['./scratchcard.component.scss']
})
export class ScratchCardComponent implements OnInit {

  routes = {
    scratchcard: `${environment.walletUrl}/redeem`
  };

  showCodeLocation: boolean = false;
  redeemCodeForm: FormGroup;
  walletList: any;
  selectedAddress: any;
  showSelector: boolean = false;

  @Input()
  showClose: any = false;

  @Output()
  doClose = new EventEmitter<any>();

  constructor(
    public dialog: MatDialog,
    private formBuilder: FormBuilder,
    private httpClient: HttpClient,
    private settingsService: SettingsService
  ) { }

  ngOnInit() {

    this.redeemCodeForm = this.formBuilder.group({
        address: ['', Validators.required],
        code: ['', Validators.pattern('[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}')],
        recaptcha: ['', Validators.required]
    });

    this.settingsService.getWalletList().then(wallets => {

      let promises = [];
      for (let i = 0; i < wallets.length; i++)
      {
        let wallet = wallets[i];
        promises.push(this.settingsService.getWalletDetails(wallet.label, null));
      }
      Promise.all(promises).then(wallets => {


        let list = [];

        for (let i = 0; i < wallets.length; i++)
        {
          let wallet = wallets[i];

          for (let x = 0; x < wallet.entries.length; x++)
          {
            let entry = wallet.entries[x];

            list.push({label: wallet.meta.label, address: entry.address});
          }
        }

        this.walletList = list;
      }, 
      error => {
        alert(error);
      })

    }).catch(function(err){
      console.log(err);
    });
  }

  setAddress() {
    console.log(this.selectedAddress)
    this.redeemCodeForm.controls['address'].setValue(this.selectedAddress.address);
  }

  doSubmit() {
    this.httpClient.post(this.routes.scratchcard, {
      "code": this.redeemCodeForm.value.code,
      "address": this.redeemCodeForm.value.address,
      "recaptcha": this.redeemCodeForm.value.recaptcha,
    }, this.getOptions()).subscribe(
      (data: any) => {
      alert("Code successfully redeemed.  Please refresh the wallet with the address: " + this.redeemCodeForm.value.address)
      }, // success path
      error => {
        alert(error.error);
      } // error path
    );
  }

  close() {
    this.doClose.emit()
  }

  private getOptions() {
    const headers = {
      'Content-Type': 'application/x-www-form-urlencoded',
    };

    return { headers: headers };
  }
}
