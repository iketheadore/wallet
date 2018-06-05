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

  constructor(private dialogRef:MatDialogRef<SettingsComponent>, 
              private renderer: Renderer2,
              private settingsService: SettingsService,
              private el: ElementRef) { }

 
  ngOnInit() {
  }

  doClose() {
    this.dialogRef.close();
  }

  doBackup() {

    //Get the wallet data
    this.settingsService.getBackupFile().then(wallets => {
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
             let complete = 0;

             if (backup)
             {
               for (let i = 0; i < backup.length; i++)
               {
                 let wallet = backup[i];

                 let params = {
                  label: wallet.label,
                  seed: wallet.seed,
                  aCount: wallet.aCount,
                  encrypted: wallet.encrypted
                };
                __this.settingsService.restoreSeed(params).subscribe((result: any) => { 
                  let refresh_event = new CustomEvent('refreshButtonClick', { cancelable: true, detail: {} });
                  document.dispatchEvent(refresh_event);
                  complete = complete + 1;
                  if (complete == backup.length)
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

  doRestore() {
    let params = {
      label: this.restore_name,
      seed: this.restore_seed,
      aCount: 1,
      encrypted: false
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
