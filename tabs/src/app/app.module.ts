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
import { AppComponent } from './app.component';
import { SafePipe } from './game/safe.pipe';
import { MatDialogModule } from '@angular/material/dialog';

@NgModule({
  declarations: [
    AppComponent,
    GameComponent,
    ErrorScreenComponent,
    ConnectionStatusComponent,
    SettingsComponent,
    SafePipe
  ],
  entryComponents: [
    SettingsComponent
  ],
  imports: [
  	HttpClientModule,
    BrowserModule,
    MarketplaceAppModule,
    WalletAppModule,
    MatDialogModule
  ],
  providers: [
  	ScoreboardService,
    ErrorScreenService,
    ConnectionStatusService
  ],
  bootstrap: [AppComponent]
})
export class AppModule { }
