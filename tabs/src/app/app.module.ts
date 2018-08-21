import { BrowserModule } from '@angular/platform-browser';
import { NgModule } from '@angular/core';
import { HttpClientModule } from '@angular/common/http';
import { MarketplaceAppModule } from 'kittycash-marketplace-lib';
import { WalletAppModule } from 'wallet-lib';
import { GameComponent } from "./game/game.component";
import { ScoreboardService } from "./game/scoreboard.service";
import { ErrorScreenComponent } from "./error_screen/error_screen.component";
import { ErrorScreenService } from "./error_screen/error_screen.service";
import { ConnectionStatusComponent } from "./connection_status/connection_status.component";
import { ConnectionStatusService } from "./connection_status/connection_status.service";
import { SettingsComponent } from "./settings/settings.component";
import { ScratchCardComponent } from "./scratchcard/scratchcard.component";
import { ScratchCardDialogComponent } from "./scratchcard_dialog/scratchcard_dialog.component";
import { AppComponent } from './app.component';
import { SafePipe } from './game/safe.pipe';
import { SettingsService } from "./settings/settings.service";
import { MatDialogModule } from '@angular/material/dialog';
import { FormsModule, ReactiveFormsModule } from '@angular/forms';
import { RECAPTCHA_SETTINGS, RecaptchaModule, RecaptchaSettings } from 'ng-recaptcha';
import { RecaptchaFormsModule } from 'ng-recaptcha/forms';
import { environment } from '../environments/environment';
import { NgxMaskModule } from 'ngx-mask'

@NgModule({
  declarations: [
    AppComponent,
    GameComponent,
    ErrorScreenComponent,
    ConnectionStatusComponent,
    SettingsComponent,
    ScratchCardComponent,
    ScratchCardDialogComponent,
    SafePipe
  ],
  entryComponents: [
    SettingsComponent,
    ScratchCardDialogComponent
  ],
  imports: [
  	HttpClientModule,
    BrowserModule,
    MarketplaceAppModule,
    WalletAppModule,
    MatDialogModule,
    FormsModule,
    ReactiveFormsModule,
    RecaptchaModule.forRoot(),
    RecaptchaFormsModule,
    NgxMaskModule.forRoot()
  ],
  providers: [
  	ScoreboardService,
    ErrorScreenService,
    ConnectionStatusService,
    SettingsService,
    {
      provide: RECAPTCHA_SETTINGS,
      useValue: {siteKey: environment.recaptchaSiteKey} as RecaptchaSettings,
    }
  ],
  bootstrap: [AppComponent]
})
export class AppModule { }
