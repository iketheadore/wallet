import { Component, OnInit, Directive, ElementRef, HostListener, HostBinding, Renderer, Input, Renderer2} from '@angular/core';
import { finalize } from 'rxjs/operators';
import { MatDialogRef } from '@angular/material';
import { SettingsService } from './settings.service';
import { FormsModule } from '@angular/forms';

@Component({
  selector: 'app-settings',
  templateUrl: './settings.component.html',
  styleUrls: ['./settings.component.scss']
})
export class SettingsComponent implements OnInit {
 
  restore_name: string;
  restore_seed: string;
  wallets_list: Array<any> = [];
  restore_wallets_list: any = false;

  constructor(private dialogRef:MatDialogRef<SettingsComponent>, 
              private renderer: Renderer2,
              private settingsService: SettingsService,
              private el: ElementRef) { }

 
  ngOnInit() {

    this.refreshWalletList();
   
  }

  refreshWalletList() {
    this.settingsService.getWalletList().then(wallets_list => {
      this.wallets_list = wallets_list;
    }).catch(function(err){
      console.log(err);
    });
  }
  doClose() {
    this.dialogRef.close();
  }

  unlockWallet(wallet: any)
  {
    let __this = this;
    this.settingsService.getWalletDetails(wallet.label, wallet.password).then(function(success){
      __this.refreshWalletList();
    }).catch(function(err){
       alert(err);
    });
  }

  walletsToRestore() {

    let success = false;
    let missing_password = false;

    for (let i = 0; i < this.restore_wallets_list.length; i++)
    {
      if (this.restore_wallets_list[i].restore)
      {
         success = true;
      }
    }

    for (let i = 0; i < this.restore_wallets_list.length; i++)
    {
      if (this.restore_wallets_list[i].restore && this.restore_wallets_list[i].encrypted && (!this.restore_wallets_list[i].password || (this.restore_wallets_list[i].password && this.restore_wallets_list[i].password.length <= 0)))
      {
        return false;
      }
    }

    return success;
  }
  walletsToBackup() {
    for (let i = 0; i < this.wallets_list.length; i++)
    {
      if (this.wallets_list[i].backup)
      {
         return true;
      }
    }

    return false;
  }
  doBackup() {

    let wallets = [];

    for (let i = 0; i < this.wallets_list.length; i++)
    {
      if (this.wallets_list[i].backup)
      {
        wallets.push(this.wallets_list[i]);
      }
    }
    //Get the wallet data
    this.settingsService.getBackupFile(wallets).then(wallets => {
      //Now prepare the data in our backup format.

      let data = [];
      for (let i = 0; i < wallets.length; i++)
      {
        let wallet = wallets[i];
        if (wallet)
        {
          let obj = {
            label: wallet.meta.label, 
            seed: wallet.meta.seed, 
            version: wallet.meta.version, 
            aCount: wallet.entry_count, 
            encrypted: wallet.meta.encrypted
          };
          data.push(obj);
        } 
      }

      var dataStr = "data:text/json;charset=utf-8," + encodeURIComponent(JSON.stringify(data));
      var downloadAnchorNode = document.createElement('a');
      downloadAnchorNode.setAttribute("href",     dataStr);
      downloadAnchorNode.setAttribute("download", "kittycash-wallet.json");
      downloadAnchorNode.click();
      downloadAnchorNode.remove();
    });    
  }

  doRestoreBackup() {

    let __this = this;

    let inputEl = this.el.nativeElement.querySelector("#restoreBackupFile");

    if (inputEl.files.length == 0) return;

    let files :FileList = inputEl.files;
    
    if (files.length > 0)
    {
      let file:File = files[0];
      let reader:FileReader = new FileReader();

      reader.onloadend = function(e){
        if (reader.result && reader.result.length > 0)
        {
          //Try to parse the json

           try {
             let backup = JSON.parse(reader.result);

             __this.restore_wallets_list = backup;

           } catch(e) {
             alert("Invalid backup file.");
           }
        }
        else
        {
          alert("Invalid restore file");
        }
      }

      reader.readAsText(file);
    }
    
  }

  restoreSelectedBackups()
  {
     let __this = this;
     let complete = 0;

     let wallets = [];

    for (let i = 0; i < this.restore_wallets_list.length; i++)
    {
      if (this.restore_wallets_list[i].restore)
      {
        wallets.push(this.restore_wallets_list[i]);
      }
    }

     if (wallets.length > 0)
     {
       for (let i = 0; i < wallets.length; i++)
       {
         let wallet = wallets[i];

         let params = {
          label: wallet.label,
          seed: wallet.seed,
          aCount: wallet.aCount,
          encrypted: wallet.encrypted,
          password: null
        };

        if (params.encrypted)
        {
          params.password = wallet.password;
        }

        __this.settingsService.restoreSeed(params).subscribe((result: any) => { 
          let refresh_event = new CustomEvent('refreshButtonClick', { cancelable: true, detail: {} });
          document.dispatchEvent(refresh_event);
          complete = complete + 1;
          if (complete == wallets.length)
          {
             __this.dialogRef.close();
          }

        });
       }
     }
     else
     {
       alert("Invalid backup file");
     }
  }

  doRestore() {
    let params = {
      label: this.restore_name,
      seed: this.restore_seed,
      aCount: 1,
      encrypted: false,
      password: null
    };
    this.settingsService.restoreSeed(params).subscribe((result: any) => { 
      let refresh_event = new CustomEvent('refreshButtonClick', { cancelable: true, detail: {} });
      document.dispatchEvent(refresh_event);
      this.dialogRef.close();
    });
  }

  toggleFrame() {
    if (document.body.classList.contains('colored-frame'))
    {
      this.renderer.removeClass(document.body, 'colored-frame');
    }
    else
    {
      this.renderer.addClass(document.body, 'colored-frame');
    }
    
  }

}
