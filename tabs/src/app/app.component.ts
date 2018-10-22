import { Component, OnInit, HostListener, ViewChild } from '@angular/core';
import { Http } from '@angular/http';
import { Observable } from 'rxjs/Observable';
import 'rxjs/add/operator/map'
import 'rxjs/add/operator/catch'
import { ErrorScreenService } from './error_screen/error_screen.service';
import { MatDialog } from '@angular/material';
import { SettingsComponent } from './settings/settings.component';
import { ScratchCardDialogComponent } from './scratchcard_dialog/scratchcard_dialog.component';
import { WalletAppModule } from 'wallet-lib';

@Component({
  selector: 'app-root',
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.scss']
})
export class AppComponent implements OnInit {


  version: string;
  releaseVersion: string;
  updateAvailable: boolean;
  currentTab = 'wallet';
  settingsDialog: any;
  tourSteps: any = [
    {
      content: "This is the <strong>play</strong> button.<br>It uses HTML!",
      element: "play_button",
      position: 'left',
      background: 'yellow',
      size: {
        width: 200,
        height: 50
      },
      padding: {
        x: 0,
        y: 10
      },
      pre_script: null
    },
    {
      content: "This is a full screen dialog.  Again we can embed anything we like:<br><br><img class='img-fluid' src='https://i.ytimg.com/vi/E9U9xS4thxU/hqdefault.jpg'>",
      element: null,
      position: 'full',
      background: 'yellow',
      size: {
        width: 500,
        height: 500
      },
      padding: {
        x: 0,
        y: 0
      },
      pre_script: null
    },
    {
      content: "This step can open and close a dialog or page with some simple code.",
      element: "restore_from_file",
      position: 'top',
      background: 'purple',
      size: {
        width: 150,
        height: 100
      },
      padding: {
        x: 0,
        y: 0
      },
      pre_script: 'this.doOpenSettings()',
      post_script: 'this.doCloseSettings()'
    },
       {
      content: "This is another full screen dialog that closed the previous window.<br><br><img class='img-fluid' src='https://boygeniusreport.files.wordpress.com/2015/06/funny-cat.jpg?quality=98&strip=all&w=782'>",
      element: null,
      position: 'full',
      background: 'purple',
      size: {
        width: 500,
        height: 500
      },
      padding: {
        x: 0,
        y: 0
      },
      pre_script: null
    }
  ];

  tour: any =  {
    show: false,
    step: 0,
    x: 100,
    y: 0
  }


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


  prevStep() {
    if (this.tour.currentStep.post_script)
    {
      eval(this.tour.currentStep.post_script); 
    }
    this.tour.show = false;
    this.tour.step = this.tour.step - 1;

    this.renderTourStep();
  }
  nextStep() {
    this.tour.show = false;
    this.tour.step = this.tour.step + 1;

    this.renderTourStep();
  }
  private renderTourStep()
  {
      if (this.tour.currentStep && this.tour.currentStep.post_script)
      {
        eval(this.tour.currentStep.post_script); 
      }

     //Get the step
      this.tour.currentStep = this.tourSteps[this.tour.step];

      if (this.tour.currentStep)
      {

        if (this.tour.currentStep.pre_script)
        {
          eval(this.tour.currentStep.pre_script);
        }

        let __this = this;

        setTimeout(function(){
          var elem = document.getElementById(__this.tour.currentStep.element);

          if (elem && __this.tour.currentStep.position != "full")
          {
            var pos = __this.getPosition(elem);

            if (pos)
            {
              var x = pos.x + __this.tour.currentStep.padding.x;
              var y = pos.y + __this.tour.currentStep.padding.y;

              switch(__this.tour.currentStep.position) {
                  case 'bottom':
                      y = y + elem.clientHeight;
                      break;
                  case 'top':
                      y = y - elem.clientHeight - __this.tour.currentStep.size.height;
                      break;
                  case 'left': 
                      x = x - elem.clientWidth - __this.tour.currentStep.size.width;
                      break;
                  case 'right':
                      x = x + elem.clientWidth;

                  default:
                      //Do nothing I guess
              }
              __this.tour.x = x;
              __this.tour.y = y;
              __this.tour.show = true;
            }
          }
          else
          {
            __this.tour.x = (window.innerWidth / 2) - (__this.tour.currentStep.size.width / 2);
            __this.tour.y = (window.innerHeight / 2) - (__this.tour.currentStep.size.height / 2);
            __this.tour.show = true;
          }
        }, 250);
      }
  }

  startTour() {
      this.tour.step = 0;
      this.renderTourStep();
  } 

  closeTour() {
    this.tour.step = 0;
    this.tour.show = false;
  }

  getPosition(el) {
    var xPos = 0;
    var yPos = 0;

    while (el) {
      if (el.tagName == "BODY") {
        // deal with browser quirks with body/window/document and page scroll
        var xScroll = el.scrollLeft || document.documentElement.scrollLeft;
        var yScroll = el.scrollTop || document.documentElement.scrollTop;

        xPos += (el.offsetLeft - xScroll + el.clientLeft);
        yPos += (el.offsetTop - yScroll + el.clientTop);
      } else {
        // for all other non-BODY elements
        xPos += (el.offsetLeft - el.scrollLeft + el.clientLeft);
        yPos += (el.offsetTop - el.scrollTop + el.clientTop);
      }

      el = el.offsetParent;
    }

    return {
      x: xPos,
      y: yPos
    };
   }


  ngOnInit() {

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

  @HostListener('document:walletCreated', ['$event'])
    onCreation(ev:any) {
      ev.preventDefault();
      let __this = this;
      setTimeout(function(){
        if (!localStorage.getItem('do_not_show_scratchcard'))
        {  
          localStorage.setItem('do_not_show_scratchcard', 'true');

          __this.dialog.open(ScratchCardDialogComponent, { width: '700px' });
        }
      }, 1000);
      
  }

  doRefresh() {
    let event = new CustomEvent('refreshButtonClick', { cancelable: true, detail: {} });
    document.dispatchEvent(event);
  }

  doOpenSettings(){
    this.settingsDialog = this.dialog.open(SettingsComponent, { width: '700px' });
  }

  doCloseSettings(){
   this.settingsDialog.close();
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